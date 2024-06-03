package frxml

import (
	"strings"
	"os"
	"io"
	"fmt"
)

const (
	startNode int = iota
	attributeNode
	endNode
	autoNode
	textNode
	error
)

type KeyValue struct {
	Key string
	Value string
}

type Node struct {
	nodeType int
	Value string
	attribute KeyValue
	children []*Node
	parent *Node
}

type XmlDocument struct {
	Declare map[string]map[string]string
	DefaultNamespace string
	Namespace map[string]string
	root *Node
}

func (document *XmlDocument) lexerXml(xmlText string, fn func(node *Node)) {
	status := 0
	key := ""
	value := ""
	type_ := ""
	tag := ""

	for _, ch := range xmlText {
		if ch == '\r' || ch == '\n' {
			continue
		}
		switch status {
		case 0:
			if ch == '<' {
				status = 1
				value = strings.TrimSpace(value)
				if value != "" {
					n := new(Node)
					n.nodeType = textNode
					n.Value = value
					fn(n)
					value = ""
				}
			} else if ch == '&' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else {
				// 处理xml文本节点
				value += string(ch)
			}
		case 1:
			if ch == '?' {
				// 处理xml声明
				status = 9
			} else if ch == '!' {
				// 处理xml注释
				status = 14
			} else if ch == '/' {
				// 处理xml闭合标签
				status = 7
			} else if ch == '<' || ch == '>' || ch == '&' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch != ' ' && ch != '\t' {
				// 处理xml开始节点
				tag += string(ch)
				status = 2
			}
		case 2:
			if ch == '/' {
				// 处理xml自闭合标签
				status = 6
			} else if ch == '<' || ch == '&' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch == '>' {
				// 开始标签结束
				status = 0
				n := new(Node)
				n.nodeType = startNode
				n.Value = tag
				fn(n)
				tag = ""
			} else if ch != ' ' && ch != '\t' {
				// 处理xml开始节点
				tag += string(ch)
			} else {
				status = 3
			}
		case 3:
			if ch == '/' {
				// 处理xml自闭合标签
				status = 6
			} else if ch == '<' || ch == '&' || ch == '=' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch == '>' {
				// 开始标签结束
				n := new(Node)
				n.nodeType = startNode
				n.Value = tag
				fn(n)
				tag = ""
				status = 0
			} else if ch != ' ' && ch != '\t' {
				// 处理xml属性节点
				key += string(ch)
				status = 19
			}
		case 4:
			if ch == '"' {
				// 处理xml属性节点
				status = 5
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 5:
			if ch == '"' {
				// 处理xml属性节点
				n := new(Node)
				n.nodeType = attributeNode
				n.attribute = KeyValue {key, value}
				fn(n)
				status = 3
				key = ""
				value = ""
			} else if ch == '<' || ch == '&' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else {
				value += string(ch)
			}
		case 6:
			// 自闭合标签结束
			if ch == '>' {
				status = 0
				n := new(Node)
				n.nodeType = autoNode
				n.Value = tag
				fn(n)
				tag = ""
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 7:
			// 处理闭合标签
			if ch == '!' || ch == '/' || ch == '>' || ch == '<' || ch == '&' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch != ' ' && ch != '\t' {
				// 处理xml闭合标签
				tag += string(ch)
				status = 8
			}
		case 8:
			// 处理闭合标签
			if ch == '!' || ch == '/' || ch == '<' || ch == '&' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch == '>' {
				// 闭合标签结束
				status = 0
				n := new(Node)
				n.nodeType = endNode
				n.Value = tag
				fn(n)
				tag = ""
			} else if ch != ' ' && ch != '\t' {
				// 处理xml闭合标签
				tag += string(ch)
				status = 8
			}
		case 9:
			if ch == '?' {
				status = 13
			} else if ch == '!' || ch == '/' || ch == '<' || ch == '&' || ch == '>' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch != ' ' {
				type_ += string(ch)
			} else {
				_, ok := document.Declare[type_]
				if !ok {
					document.Declare[type_] = make(map[string]string)
				}
				status = 10
			}
		case 10:
			if ch == '?' {
				status = 13
			} else if ch == '>' || ch == '!' || ch == '/' || ch == '<' || ch == '&' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch != '=' {
				key += string(ch)
				status = 11
			}
		case 11:
			if ch == '>' || ch == '!' || ch == '/' || ch == '<' || ch == '&' || ch == '?' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else if ch != '=' && ch != ' ' && ch != '\t' {
				key += string(ch)
			} else if ch == '=' {
				status = 12
			}
		case 12:
			if ch == '"' {
				status = 20
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 13:
			if ch == '>' {
				status = 0
				type_ = ""
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 14:
			if ch == '-' {
				status = 15
			} else if ch == '[' {
				status = 21
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 15:
			if ch == '-' {
				status = 16
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 16:
			if ch == '-' {
				status = 17
			}
		case 17:
			if ch == '-' {
				status = 18
			} else {
				status = 16
			}
		case 18:
			if ch == '>' {
				status = 0
			} else {
				status = 16
			}
		case 19:
			// 处理xml属性节点
			if ch == '=' {
				status = 4
			} else if ch == '<' || ch == '&' || ch == '=' || ch == '>' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else {
				key += string(ch)
			}
		case 20:
			if ch == '"' {
				document.Declare[type_][strings.TrimSpace(key)] = value
				key = ""
				value = ""
				status = 10
			} else if ch == '<' || ch == '&' || ch == '=' || ch == '>' || ch == ' ' || ch == '\t' {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			} else {
				value += string(ch)
			}
		case 21:
			if ch == 'C' || ch == 'c' {
				status = 22
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 22:
			if ch == 'D' || ch == 'd' {
				status = 23
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 23:
			if ch == 'A' || ch == 'a' {
				status = 24
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 24:
			if ch == 'T' || ch == 't' {
				status = 25
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 25:
			if ch == 'A' || ch == 'a' {
				status = 26
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 26:
			if ch == '[' {
				status = 27
			} else {
				n := new(Node)
				n.nodeType = error
				n.Value = "词法分析错误：无法解析的字符"
				fn(n)
			}
		case 27:
			if ch == ']' {
				status = 28
			} else {
				value += string(ch)
			}
		case 28:
			if ch == ']' {
				status = 29
			} else {
				value += "]" + string(ch)
				status = 27
			}
		case 29:
			if ch == '>' {
				status = 0
			} else {
				value += "]]" + string(ch)
				status = 27
			}
		}
	}
}

func (document *XmlDocument) ParseXml(xmlText string) {
	var nowNode *Node
	attributeNodes := make([]*Node, 0)

	document.lexerXml(xmlText, func(node *Node) {
		switch node.nodeType {
		case attributeNode:
			if strings.HasPrefix(node.attribute.Key, "xmlns") {
				if strings.Contains(node.attribute.Key, ":") {
					document.Namespace[strings.Split(node.attribute.Key, ":")[1]] = node.attribute.Value
				} else {
					document.DefaultNamespace = node.attribute.Value
				}
			}
			attributeNodes = append(attributeNodes, node)
		case startNode:
			if nowNode == nil {
				if document.root != nil {
					fmt.Println("只能有一个根标签")
					return
				}
				nowNode = node
				document.root = nowNode
			} else {
				node.parent = nowNode
				nowNode.children = append(nowNode.children, node)
				nowNode = node
			}
			nowNode.children = append(nowNode.children, attributeNodes...)
			attributeNodes = make([]*Node, 0)
		case autoNode:
			if nowNode == nil {
				document.root = node
			} else {
				node.parent = nowNode
				nowNode.children = append(nowNode.children, node)
			}
			node.children = append(node.children, attributeNodes...)
			attributeNodes = make([]*Node, 0)
		case endNode:
			if nowNode == nil {
				fmt.Println("语法错误，标签不匹配")
				return
			}
			if node.Value == nowNode.Value {
				nowNode = nowNode.parent
			} else {
				fmt.Println("语法错误，标签不匹配")
			}
		case textNode:
			if nowNode == nil {
				fmt.Println("语法错误，文本不在标签内")
				return
			}
			node.parent = nowNode
			nowNode.children = append(nowNode.children, node)
		case error:
			fmt.Println(node.Value)
		}
	})
	if nowNode != nil {
		fmt.Println("语法错误，根标签不匹配")
	}
}

func NewXmlDocument() (xmlDocument XmlDocument) {
	xmlDocument = XmlDocument { Declare: make(map[string]map[string]string), Namespace: make(map[string]string) }
	return xmlDocument
}

func (document *XmlDocument) LoadXmlFromFile(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		fmt.Println(err)
	} else {
		defer file.Close()
		buffer := make([]byte, 1024)
		xmlText := ""
		for ;; {
			n, err2 := file.Read(buffer)
			if err2 != nil {
				if err2 == io.EOF {
					xmlText += string(buffer[:n])
					break
				}
				fmt.Println(err2)
				return
			}
			xmlText += string(buffer[:n])
		}
		document.ParseXml(xmlText)
	}
}

func (document *XmlDocument) Root() *Node {
	return document.root
}

func (node *Node) Children() []*Node {
	res := make([]*Node, 0)
	for _, child := range node.children {
		if child.nodeType == startNode || child.nodeType == autoNode {
			res = append(res, child)
		}
	}
	return res
}

func (node *Node) Parent() *Node {
	return node.parent
}

func (node *Node) GetText() string {
	res := ""
	for _, v := range node.children {
		if v.nodeType == textNode {
			res += v.Value
		}
	}
	return res
}

func (node *Node) SetText(text string) {
	for i := 0; i < len(node.children); i++ {
		if node.children[i].nodeType == textNode {
			node.children = append(node.children[:i], node.children[i + 1:]...)
		}
	}
	n := new(Node)
	n.Value = text
	n.nodeType = textNode
	node.children = append(node.children, n)
}

func (node *Node) Child(tag string) *Node {
	for _, n := range node.children {
		if n.Value == tag && (n.nodeType == startNode || n.nodeType == autoNode) {
			return n
		}
	}
	return nil
}

func (node *Node) GetAttributes(mapper map[string]string) {
	for _, child := range node.children {
		if child.nodeType == attributeNode {
			mapper[child.attribute.Key] = child.attribute.Value
		}
	}
}

func (node *Node) Add(name string) *Node {
	n := new(Node)
	n.nodeType = startNode
	n.Value = name
	node.children = append(node.children, n)
	return n
}

func (node *Node) Delete() {
	for i := 0; i < len(node.parent.children); i++ {
		if node.parent.children[i] == node {
			node.parent.children = append(node.parent.children[:i], node.parent.children[i + 1:]...)
			return
		}
	}
}

func (node *Node) Update(newName string) {
	node.Value = newName
}

func (node *Node) AddAttribute(key, value string) {
	attr := KeyValue {Key: key, Value: value}
	attrNode := new(Node)
	attrNode.nodeType = attributeNode
	attrNode.attribute = attr
	node.children = append(node.children, attrNode)
}

func (node *Node) DelAttribute(key string) {
	for i := 0; i < len(node.children); i++ {
		if node.children[i].nodeType == attributeNode && node.children[i].attribute.Key == key {
			node.children = append(node.children[:i], node.parent.children[i + 1:]...)
			i--
		}
	}
}

func (node *Node) UpdateAttribute(key, value string) {
	for i := 0; i < len(node.children); i++ {
		if node.children[i].nodeType == attributeNode && node.children[i].attribute.Key == key {
			node.children[i].attribute.Value = value
		}
	}
}

func (document *XmlDocument) GetXml() string {
	if document.root == nil {
		return ""
	}
	res := ""
	for tag, kv := range document.Declare {
		res += "<?" + tag
		for k, v := range kv {
			res += " " + k + "=\"" + v + "\""
		}
		res += "?>\n"
	}
	var fun func(node *Node)
	fun = func(node *Node) {
		res += "<" + node.Value
		attrs := make(map[string]string, 0)
		node.GetAttributes(attrs)
		for k, v := range attrs {
			res += " " + k + "=\"" + v + "\""
		}
		text := node.GetText()
		res += ">\n"
		if text != "" {
			res += node.GetText() + "\n"
		}
		for _, child := range node.Children() {
			fun(child)
		}
		res += "</" + node.Value + ">\n"
	}
	fun(document.root)
	return res
}

func (document *XmlDocument) Save(fileName string) {
	file, err2 := os.Open(fileName, os.O_WROBLY)
	if err2 != nil {
		fmt.Println(err2)
		return
	}
	defer file.Close()
	_, err := file.Write([]byte(document.GetXml()))
	if err != nil {
		fmt.Println(err)
	}
}