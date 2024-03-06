package binchunk

import (
	"encoding/binary"
	"math"
)

// 二进制chunk内部使用的数据类型大致可以分为数字、字符串和列表三种。

// 数字类型主要包括字节、C语言整型（后文简称cint）、C语言size_t类型（简称size_t）、Lua整数、Lua浮点数五种。其中，字节类型用来存放一些比较小的整数值，比如Lua版本号、函数的参数个数等；cint类型主要用来表示列表长度；size_t则主要用来表示长字符串长度；Lua整数和Lua浮点数则主要在常量表里出现，记录Lua脚本中出现的整数和浮点数字面量。
// 数字类型在二进制chunk里都按照固定长度存储。除字节类型外，其余四种数字类型都会占用多个字节，具体占用几个字节则会记录在头部里

// 字符串在二进制chunk里，其实就是一个字节数组。因为字符串长度是不固定的，所以需要把字节数组的长度也记录到二进制chunk里。作为优化，字符串类型又可以进一步分为短字符串和长字符串两种，具体有三种情况：
// 1）对于NULL字符串，只用0x00表示就可以了。
// 2）对于长度小于等于253（0xFD）的字符串，先使用一个字节记录长度+1，然后是字节数组。
// 3）对于长度大于等于254（0xFE）的字符串，第一个字节是0xFF，后面跟一个size_t记录长度+1，最后是字节数组。

// 在二进制chunk内部，指令表、常量表、子函数原型表等信息都是按照列表的方式存储的。具体来说也很简单，先用一个cint类型记录列表长度，然后紧接着存储n个列表元素

type reader struct {
	data []byte
}

func (r *reader) readByte() byte {
	b := r.data[0]
	r.data = r.data[1:]
	return b
}

func (r *reader) readUint32() uint32 {
	i := binary.LittleEndian.Uint32(r.data)
	r.data = r.data[4:]
	return i
}

func (r *reader) readUint64() uint64 {
	i := binary.LittleEndian.Uint64(r.data)
	r.data = r.data[8:]
	return i
}

func (r *reader) readLuaInteger() int64 {
	return int64(r.readUint64())
}

func (r *reader) readLuaNumber() float64 {
	return math.Float64frombits(r.readUint64())
}

func (r *reader) readBytes(n uint) []byte {
	bytes := r.data[:n]
	r.data = r.data[n:]
	return bytes
}

func (r *reader) readString() string {
	size := uint(r.readByte())
	if size == 0 {
		return ""
	}
	if size == 0xFF {
		size = uint(r.readUint64())
	}
	bytes := r.readBytes(size - 1) // -1 for '\0'
	return string(bytes)
}

func (r *reader) checkHeader() {
	if string(r.readBytes(4)) != LUA_SIGNATURE {
		panic("not a precompiled chunk!")
	} else if r.readByte() != LUAC_VERSION {
		panic("version mismatch!")
	} else if r.readByte() != LUAC_FORMAT {
		panic("format mismatch!")
	} else if string(r.readBytes(6)) != LUAC_DATA {
		panic("corrupted!")
	} else if r.readByte() != CINT_SIZE {
		panic("int size mismatch!")
	} else if r.readByte() != CSIZET_SIZE {
		panic("size_t size mismatch!")
	} else if r.readByte() != INSTRUCTION_SIZE {
		panic("instruction size mismatch!")
	} else if r.readByte() != LUA_INTEGER_SIZE {
		panic("lua_Integer size mismatch!")
	} else if r.readByte() != LUA_NUMBER_SIZE {
		panic("lua_Number size mismatch!")
	} else if r.readLuaInteger() != LUAC_INT {
		panic("endianness mismatch!")
	} else if r.readLuaNumber() != LUAC_NUM {
		panic("float format mismatch!")
	}
}

func (r *reader) readProto(parentSource string) *Prototype {
	source := r.readString()
	if source == "" {
		source = parentSource
	}
	return &Prototype{
		Source:          source,
		LineDefined:     r.readUint32(),
		LastLineDefined: r.readUint32(),
		NumParams:       r.readByte(),
		IsVararg:        r.readByte(),
		MaxStackSize:    r.readByte(),
		Code:            r.readCode(),
		Constants:       r.readConstants(),
		Upvalues:        r.readUpvalues(),
		Protos:          r.readProtos(source),
		LineInfo:        r.readLineInfo(),
		LocVars:         r.readLocVars(),
		UpvalueNames:    r.readUpvalueNames(),
	}
}

func (r *reader) readCode() []uint32 {
	code := make([]uint32, r.readUint32()) // 指令表大小
	for i := range code {
		code[i] = r.readUint32()
	}
	return code
}

func (r *reader) readConstants() []interface{} {
	constants := make([]interface{}, r.readUint32()) // 常量表大小
	for i := range constants {
		constants[i] = r.readConstant()
	}
	return constants
}

func (r *reader) readConstant() interface{} {
	switch r.readByte() { // tag
	case TAG_NIL:
		return nil
	case TAG_BOOLEAN:
		return r.readByte() != 0
	case TAG_INTEGER:
		return r.readLuaInteger()
	case TAG_NUMBER:
		return r.readLuaNumber()
	case TAG_SHORT_STR:
		return r.readString()
	case TAG_LONG_STR:
		return r.readString()
	default:
		panic("corrupted!")
	}
}

func (r *reader) readUpvalues() []Upvalue {
	upvalues := make([]Upvalue, r.readUint32()) // upvalue表大小
	for i := range upvalues {
		upvalues[i] = Upvalue{
			Instack: r.readByte(),
			Idx:     r.readByte(),
		}
	}
	return upvalues
}

func (r *reader) readProtos(parentSource string) []*Prototype {
	protos := make([]*Prototype, r.readUint32()) // 子函数原型表大小
	for i := range protos {
		protos[i] = r.readProto(parentSource)
	}
	return protos
}

func (r *reader) readLineInfo() []uint32 {
	lineInfo := make([]uint32, r.readUint32()) // 行号表大小
	for i := range lineInfo {
		lineInfo[i] = r.readUint32()
	}
	return lineInfo
}

func (r *reader) readLocVars() []LocVar {
	locVars := make([]LocVar, r.readUint32()) // 局部变量表大小
	for i := range locVars {
		locVars[i] = LocVar{
			VarName: r.readString(),
			StartPC: r.readUint32(),
			EndPC:   r.readUint32(),
		}
	}
	return locVars
}

func (r *reader) readUpvalueNames() []string {
	upvalueNames := make([]string, r.readUint32()) // upvalue名列表大小
	for i := range upvalueNames {
		upvalueNames[i] = r.readString()
	}
	return upvalueNames
}
