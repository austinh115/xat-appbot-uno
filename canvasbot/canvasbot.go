package canvasbot

import (
	"../config"
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type CanvasBot struct {
	chat         int
	conn         net.Conn
	incoming     chan string
	outgoing     chan string
	writer       *bufio.Writer
	reader       *bufio.Reader
	killSignal   chan bool
	HandlePacket func(bot *CanvasBot, packet *Packet)
	Config       *config.Configuration
	Players      map[string]int
	Done         bool
	LoggingIn    bool
	JoinData     *Packet
	*UNOGame
}

func (bot *CanvasBot) Open() error {
	// connect to this socket
	conn, err := net.Dial("tcp", "fwdelb00-1964376362.us-east-1.elb.amazonaws.com:10032")
	if err == nil {
		bot.conn = conn

		bot.writer = bufio.NewWriter(conn)
		bot.reader = bufio.NewReader(conn)

		// Read & write data to xat
		go func() {
			for {
				select {
				case <-bot.killSignal:
					return
				case data := <-bot.outgoing:
					log.Printf("[canvasbot][send] %s", data)

					_, err := fmt.Fprintf(bot.conn, data)
					if err != nil {
						fmt.Println(err)
					}
				}
			}
		}()

		go func() {
			for {
				select {
				case <-bot.killSignal:
					return
				default:
					line, err := bot.reader.ReadString('\000')
					if err != nil {
						if err == io.EOF {
							bot.Close()
							continue
						}
						fmt.Println("bufio.reader.ReadString failed.")
						fmt.Println(err)
						time.Sleep(100 * time.Millisecond)
						bot.killSignal <- true
						continue
					}
					log.Printf("[canvasbot][read] %s", line)
					bot.incoming <- line
				}
			}
		}()

		// Handle incoming data
		go func() {
			for {
				select {
				case <-bot.killSignal:
					return
				case data := <-bot.incoming:
					if data == "" {
						return
					}
					r := regexp.MustCompile(`([^\s]+)="([^"]+)"`)
					matches := r.FindAllStringSubmatch(data, -1)
					r = regexp.MustCompile(`<(.*?)\s`)
					tag := r.FindStringSubmatch(data)[1]

					attributes := make(map[string]string)

					for k := range matches {
						attributes[matches[k][1]] = matches[k][2]
					}

					packet := NewPacket(tag, attributes)

					bot.HandlePacket(bot, packet)
				}
			}
		}()
	}

	return err
}

func (bot *CanvasBot) Close() {
	bot.killSignal <- true
	_ = bot.conn.Close()
	log.Println("[canvasbot][connect] Disconnected from xat")
}

func (bot *CanvasBot) SendJ2(packet *Packet) {
	j2 := [][]interface{}{
		{"cb", 0},

		{"l5", 65535},
		{"l4", 500},
		{"l3", 500},
		{"l2", 0},

		{"y", packet.GetAttribute("i")},

		{"k", bot.JoinData.GetAttribute("k1")},
		{"k3", bot.JoinData.GetAttribute("k3")},

		{"z", "m1.5.12,3"},

		{"d1", bot.JoinData.GetAttribute("d1")},

		{"p", 0},
		{"c", config.Config.Chat},
		{"f", 0},

		{"u", bot.JoinData.GetAttribute("i")},
	}

	for i := 0; i <= 25; i++ {
		if i == 1 {
			continue
		}
		attr := "d" + strconv.Itoa(i)
		if bot.JoinData.HasAttribute(attr) {
			j2 = append(j2, []interface{}{attr, bot.JoinData.GetAttribute(attr)})
		}
	}

	if bot.JoinData.HasAttribute("dx") {
		j2 = append(j2, []interface{}{"dx", bot.JoinData.GetAttribute("dx")})
	}

	if bot.JoinData.HasAttribute("dt") {
		j2 = append(j2, []interface{}{"dt", bot.JoinData.GetAttribute("dt")})
	}

	j2 = append(j2, [][]interface{}{
		{"N", config.Config.BotInfo.Username},
		{"n", config.Config.BotInfo.Name},
		{"a", config.Config.BotInfo.Avatar},
		{"h", config.Config.BotInfo.Home},

		{"v", ""},
	}...)

	bot.Send("j2", j2)
}

func (bot *CanvasBot) JoinChat() {
	if bot.LoggingIn {
		bot.Send("y",
			[][]interface{}{
				{"r", 2},
				{"v", 0},
				{"u", 1001},
			},
		)
	} else {
		bot.Send("y",
			[][]interface{}{
				{"r", config.Config.Chat},
				{"v", 0},
				{"u", bot.JoinData.GetAttribute("i")},
				//{"z", 8335799305056508195},
			},
		)
	}
}

func (bot *CanvasBot) Login() {
	//<v n="username" a="api_key" />
	bot.Send("v", [][]interface{}{
		{"n", config.Config.BotInfo.Username},
		{"a", config.Config.BotInfo.ApiKey},
	})
}

func (bot *CanvasBot) GetId() string {
	return bot.JoinData.GetAttribute("i")
}

func (bot *CanvasBot) SendMessage(msg string) {
	// <m t="hi" u="1529899154" />
	bot.Send("m",
		[][]interface{}{
			{"t", "_" + msg},
			{"u", bot.JoinData.GetAttribute("i")},
		},
	)
}

func (bot *CanvasBot) Send(tag string, data [][]interface{}) {
	var buf bytes.Buffer
	buf.WriteString("<" + tag + " ")

	for k := range data {
		if data[k][0] == "" || (data[k][1] == "" && data[k][0] != "h") {
			continue
		}
		buf.WriteString(fmt.Sprintf("%v=\"%v\" ", data[k][0], data[k][1]))
	}

	buf.WriteString("/>\000")

	bot.outgoing <- buf.String()
}

func (bot *CanvasBot) GetPlayerByRegname(s string) int {
	val, ok := bot.Players[strings.ToLower(s)]
	if !ok {
		return -1
	}
	return val
}

func (bot *CanvasBot) GetRegnameById(find int) string {
	for regname, id := range bot.Players {
		if id == find {
			return strings.Title(regname)
		}
	}
	return ""
}

func New(config *config.Configuration, handler func(bot *CanvasBot, packet *Packet)) (*CanvasBot, error) {
	var bot = &CanvasBot{
		chat:         config.Chat,
		incoming:     make(chan string),
		outgoing:     make(chan string),
		killSignal:   make(chan bool),
		HandlePacket: handler,
		Config:       config,
		Players:      make(map[string]int),
		LoggingIn:    true,
	}
	return bot, nil
}
