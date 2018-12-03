package canvasbot

type Packet struct {
	Tag string
	Attributes map[string]string
}

func NewPacket(s string, strings map[string]string) *Packet {
	packet := &Packet{
		Tag: s,
		Attributes: strings,
	}
	return packet
}

func (packet Packet) HasAttribute(attribute string) bool {
	if _, ok := packet.Attributes[attribute]; ok {
		return true
	}
	return false
}

func (packet Packet) GetAttribute(attribute string) string {
	if val, ok := packet.Attributes[attribute]; ok {
		return val
	}
	return ""
}