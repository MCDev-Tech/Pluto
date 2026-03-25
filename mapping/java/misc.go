package java

import "strings"

type SingleInfo struct {
	Name      string
	Class     string
	Signature string
	Type      string
}

func Signature2Class(name string) string {
	if len(name) == 0 || strings.Contains(name, ".") {
		return name
	}
	name = strings.ReplaceAll(name, "/", ".")
	if name[0] == 'L' {
		return name[1 : len(name)-1]
	}
	return name
}

func PackClassInfo(name string) SingleInfo {
	name = Signature2Class(name)
	return SingleInfo{
		Name:      FullToClassName(name),
		Class:     name,
		Signature: ClassToByteCodeSignature(name),
		Type:      "class",
	}
}

func PackMethodInfo(name, class, signature string) SingleInfo {
	class = Signature2Class(class)
	return SingleInfo{
		Name:      name,
		Class:     class,
		Signature: signature,
		Type:      "method",
	}
}

func PackFieldInfo(name, class, signature string) SingleInfo {
	class = Signature2Class(class)
	return SingleInfo{
		Name:      name,
		Class:     class,
		Signature: signature,
		Type:      "field",
	}
}

func BuildMapping(mapping *map[SingleInfo]SingleInfo) *Mappings {
	result := Mappings{
		NotchToNamed: make(map[SingleInfo]SingleInfo, len(*mapping)),
		NamedToNotch: make(map[SingleInfo]SingleInfo, len(*mapping)),
		NotchByName:  make(map[string][]SingleInfo),
		NamedByName:  make(map[string][]SingleInfo),
	}
	for k, v := range *mapping {
		result.NotchToNamed[k] = v
		result.NamedToNotch[v] = k
		result.NotchByName[k.Name] = append(result.NotchByName[k.Name], k)
		result.NamedByName[v.Name] = append(result.NamedByName[v.Name], v)
	}
	return &result
}
