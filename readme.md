# frxml xml解析器

## 这是一个练习的项目
frxml从零开始实现词法分析，语法分析，并构建节点树，支持节点树的增加节点，删除节点，修改节点功能。
同时支持节点属性的新增、修改、删除功能
支持节点的文本变更\
不支持DTD语法\
不支持DTD和schema验证
## 使用方法
使用 NewXmlDocument 方法创建一个新的文档对象，对象类型为XmlDocument\
创建一个XmlDocument类型的变量document
```go
document := NewXmlDocument()
```
从xml文本中解析文档
```go
document.ParseXml(xmlText)  // xmlText为string类型
```
从xml文件中解析文档
```go
document.LoadXmlFromFile(fileName)
```
获取xml文档根节点
```go
node := document.Root()  // *Node
```
获取节点的所有子节点
```go
nodes := node.Children()
```
获取父节点
```go
parent := node.Parent()
```
获取标签内容为tag的子节点
```go
cnode := node.Child("tag")
```
获取节点文本
```go
text := node.GetText()
```
设置节点文本
```go
node.SetText(text)
```
获取节点所有属性
```go
attrs := make(map[string]string, 0)
node.GetAttributes(attrs)
```
添加标签内容为tag的新节点
```go
node.Add(tag)
```
删除当前节点
```go
node.Delete()
```
修改节点标签内容为tag
```go
node.Update(tag)
```
为节点新增属性
```go
node.AddAttribute("key", "value")
```
删除节点指定key的属性
```go
node.DelAttribute("key")
```
修改节点指定key的属性为value1
```go
node.UpdateAttribute("key", "value1")
```
由节点树反推xml文档
```go
fmt.Println(document.GetXml())
```
获取默认命名空间
```go
document.DefaultNamespace  // string
```
获取所有命名空间
```go
document.Namespace  // map[string][string]
```
将文档树保存到xml文件中
```go
document.Save("xmlPath")
```
### xml文档声明
如果xml文档中有如下声明
```xml
<?xml version="1.0" encoding="ISO-8859-1"?>
```
```go
document.Declare["xml"]["version"]  // 获取version
document.Declare["xml"]["encoding"]  // 获取encoding
```