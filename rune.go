package vt

// insert 向指定位置插入元素
func insert(data []rune, index int, val rune) []rune {
	if len(data) < index {
		return data
	}
	data = append(data, 0)
	copy(data[index+1:], data[index:])
	data[index] = val
	return data
}

// remove 从某个位置开始删除n个元素
func remove(data []rune, index, num int) (result []rune) {
	if index < len(data) {
		result = append(result, data[0:index]...)
	}
	if index+num < len(data) {
		result = append(result, data[index+num:]...)
	}
	return result
}

/** C0 控制字符
00000000	0	00	NUL (NULL)	空字符
00000001	1	01	SOH (Start Of Headling)	标题开始
00000010	2	02	STX (Start Of Text)	正文开始
00000011	3	03	ETX (End Of Text)	正文结束
00000100	4	04	EOT (End Of Transmission)	传输结束
00000101	5	05	ENQ (Enquiry)	请求
00000110	6	06	ACK (Acknowledge)	回应/响应/收到通知
00000111	7	07	BEL (Bell)	响铃
00001000	8	08	BS (Backspace)	退格
00001001	9	09	HT (Horizontal Tab)	水平制表符
00001010	10	0A	LF/NL(Line Feed/New Line)	换行键
00001011	11	0B	VT (Vertical Tab)	垂直制表符
00001100	12	0C	FF/NP (Form Feed/New Page)	换页键
00001101	13	0D	CR (Carriage Return)	回车键
00001110	14	0E	SO (Shift Out)	不用切换
00001111	15	0F	SI (Shift In)	启用切换
00010000	16	10	DLE (Data Link Escape)	数据链路转义
00010001	17	11	DC1/XON (Device Control 1/Transmission On)	设备控制1/传输开始
00010010	18	12	DC2 (Device Control 2)	设备控制2
00010011	19	13	DC3/XOFF (Device Control 3/Transmission Off)	设备控制3/传输中断
00010100	20	14	DC4 (Device Control 4)	设备控制4
00010101	21	15	NAK (Negative Acknowledge)	无响应/非正常响应/拒绝接收
00010110	22	16	SYN (Synchronous Idle)	同步空闲
00010111	23	17	ETB (End of Transmission Block)	传输块结束/块传输终止
00011000	24	18	CAN (Cancel)	取消
00011001	25	19	EM (End of Medium)	已到介质末端/介质存储已满/介质中断
00011010	26	1A	SUB (Substitute)	替补/替换
00011011	27	1B	ESC (Escape)	逃离/取消
00011100	28	1C	FS (File Separator)	文件分割符
00011101	29	1D	GS (Group Separator)	组分隔符/分组符
00011110	30	1E	RS (Record Separator)	记录分离符
00011111	31	1F	US (Unit Separator)	单元分隔符
01111111	127	7F	DEL (Delete)	删除
*/
func isC0Sequence(code rune) bool {
	return (code >= 0 && code <= 31) || code == 127
}

/** CSI 序列

组成部分	字符范围	ASCII
参数字节	0x30–0x3F	0–9:;<=>?
中间字节	0x20–0x2F	空格、!"#$%&'()*+,-./
最终字节	0x40–0x7E	@A–Z[\]^_`a–z{|}~

所有常见的序列都只是把参数用作一系列分号分隔的数字，如1;2;3。缺少的数字视为0（如1;;3相当于中间的数字是0，ESC[m这样没有参数的情况相当于参数为0）。某些序列（如CUU）把0视为1，以使缺少参数的情况下有意义:F.4.2。

一部分字符定义是“私有”的，以便终端制造商可以插入他们自己的序列而不与标准相冲突。包括参数字节<=>?的使用，或者最终字节0x70–0x7F（p–z{|}~）例如VT320序列CSI?25h和CSI?25l的作用是打开和关闭光标的显示。

当CSI序列含有超出0x20–0x7E范围的字符时，其行为是未定义的。这些非法字符包括C0控制字符（范围0–0x1F）、DEL（0x7F），以及高位字节。
*/
func isCSISequence(code rune) bool {
	return code >= 64 && code <= 126
}
