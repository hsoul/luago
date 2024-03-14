package codegen

import (
	. "luago/compiler/ast"
	. "luago/compiler/lexer"

	. "luago/vm"
)

var arithAndBitwiseBinops = map[int]int{
	TOKEN_OP_ADD:  OP_ADD,
	TOKEN_OP_SUB:  OP_SUB,
	TOKEN_OP_MUL:  OP_MUL,
	TOKEN_OP_MOD:  OP_MOD,
	TOKEN_OP_POW:  OP_POW,
	TOKEN_OP_DIV:  OP_DIV,
	TOKEN_OP_IDIV: OP_IDIV,
	TOKEN_OP_BAND: OP_BAND,
	TOKEN_OP_BOR:  OP_BOR,
	TOKEN_OP_BXOR: OP_BXOR,
	TOKEN_OP_SHL:  OP_SHL,
	TOKEN_OP_SHR:  OP_SHR,
}

type upvalInfo struct {
	locVarSlot int
	upvalIndex int
	index      int
}

type locVarInfo struct {
	prev     *locVarInfo
	name     string
	scopeLv  int
	slot     int
	startPC  int
	endPC    int
	captured bool
}

type funcInfo struct {
	parent    *funcInfo              // 父函数信息
	subFuncs  []*funcInfo            // 子函数信息
	usedRegs  int                    // 已经分配的寄存器数量
	maxRegs   int                    // 需要的最大寄存器数量
	scopeLv   int                    // scopeLv字段记录当前作用域层次； 作用域层次从0开始，每进入一个作用域就加1
	locVars   []*locVarInfo          // 按顺序记录函数内部声明的全部局部变量
	locNames  map[string]*locVarInfo // 记录当前生效的局部变量
	upvalues  map[string]upvalInfo
	constants map[interface{}]int // 常量表
	breaks    [][]int
	insts     []uint32
	lineNums  []uint32
	line      int
	lastLine  int
	numParams int
	isVararg  bool
}

func newFuncInfo(parent *funcInfo, fd *FuncDefExp) *funcInfo {
	return &funcInfo{
		parent:    parent,
		subFuncs:  []*funcInfo{},
		locVars:   make([]*locVarInfo, 0, 8),
		locNames:  map[string]*locVarInfo{},
		upvalues:  map[string]upvalInfo{},
		constants: map[interface{}]int{},
		breaks:    make([][]int, 1),
		insts:     make([]uint32, 0, 8),
		lineNums:  make([]uint32, 0, 8),
		line:      fd.Line,
		lastLine:  fd.LastLine,
		numParams: len(fd.ParList),
		isVararg:  fd.IsVararg,
	}
}

/* constants */

func (f *funcInfo) indexOfConstant(k interface{}) int {
	if idx, found := f.constants[k]; found {
		return idx
	}

	idx := len(f.constants)
	f.constants[k] = idx
	return idx
}

/* registers */

func (f *funcInfo) allocReg() int {
	f.usedRegs++
	if f.usedRegs >= 255 {
		panic("function or expression needs too many registers")
	}
	if f.usedRegs > f.maxRegs {
		f.maxRegs = f.usedRegs
	}
	return f.usedRegs - 1
}

func (f *funcInfo) freeReg() {
	if f.usedRegs <= 0 {
		panic("usedRegs <= 0 !")
	}
	f.usedRegs--
}

func (f *funcInfo) allocRegs(n int) int {
	if n <= 0 {
		panic("n <= 0 !")
	}
	for i := 0; i < n; i++ {
		f.allocReg()
	}
	return f.usedRegs - n
}

func (f *funcInfo) freeRegs(n int) {
	if n < 0 {
		panic("n < 0 !")
	}
	for i := 0; i < n; i++ {
		f.freeReg()
	}
}

/* lexical scope */

func (f *funcInfo) enterScope(breakable bool) {
	f.scopeLv++
	if breakable {
		f.breaks = append(f.breaks, []int{})
	} else {
		f.breaks = append(f.breaks, nil)
	}
}

func (f *funcInfo) exitScope(endPC int) {
	pendingBreakJmps := f.breaks[len(f.breaks)-1]
	f.breaks = f.breaks[:len(f.breaks)-1]

	a := f.getJmpArgA()
	for _, pc := range pendingBreakJmps {
		sBx := f.pc() - pc
		i := (sBx+MAXARG_sBx)<<14 | a<<6 | OP_JMP
		f.insts[pc] = uint32(i)
	}

	f.scopeLv--
	for _, locVar := range f.locNames {
		if locVar.scopeLv > f.scopeLv { // out of scope
			locVar.endPC = endPC
			f.removeLocVar(locVar)
		}
	}
}

func (f *funcInfo) removeLocVar(locVar *locVarInfo) {
	f.freeReg()
	if locVar.prev == nil {
		delete(f.locNames, locVar.name)
	} else if locVar.prev.scopeLv == locVar.scopeLv {
		f.removeLocVar(locVar.prev)
	} else {
		f.locNames[locVar.name] = locVar.prev
	}
}

func (f *funcInfo) addLocVar(name string, startPC int) int {
	newVar := &locVarInfo{
		name:    name,
		prev:    f.locNames[name],
		scopeLv: f.scopeLv,
		slot:    f.allocReg(),
		startPC: startPC,
		endPC:   0,
	}

	f.locVars = append(f.locVars, newVar)
	f.locNames[name] = newVar

	return newVar.slot
}

func (f *funcInfo) slotOfLocVar(name string) int {
	if locVar, found := f.locNames[name]; found {
		return locVar.slot
	}
	return -1
}

func (f *funcInfo) addBreakJmp(pc int) {
	for i := f.scopeLv; i >= 0; i-- {
		if f.breaks[i] != nil { // breakable
			f.breaks[i] = append(f.breaks[i], pc)
			return
		}
	}

	panic("<break> at line ? not inside a loop!")
}

/* upvalues */

func (f *funcInfo) indexOfUpval(name string) int {
	if upval, ok := f.upvalues[name]; ok {
		return upval.index
	}
	if f.parent != nil {
		if locVar, found := f.parent.locNames[name]; found { // 先查找局部变量
			idx := len(f.upvalues)
			f.upvalues[name] = upvalInfo{locVar.slot, -1, idx}
			locVar.captured = true
			return idx
		}
		if uvIdx := f.parent.indexOfUpval(name); uvIdx >= 0 { // 再查找upvalue
			idx := len(f.upvalues)
			f.upvalues[name] = upvalInfo{-1, uvIdx, idx}
			return idx
		}
	}
	return -1
}

func (f *funcInfo) closeOpenUpvals(line int) {
	a := f.getJmpArgA()
	if a > 0 {
		f.emitJmp(line, a, 0)
	}
}

func (f *funcInfo) getJmpArgA() int {
	hasCapturedLocVars := false
	minSlotOfLocVars := f.maxRegs
	for _, locVar := range f.locNames {
		if locVar.scopeLv == f.scopeLv {
			for v := locVar; v != nil && v.scopeLv == f.scopeLv; v = v.prev {
				if v.captured {
					hasCapturedLocVars = true
				}
				if v.slot < minSlotOfLocVars && v.name[0] != '(' {
					minSlotOfLocVars = v.slot
				}
			}
		}
	}
	if hasCapturedLocVars {
		return minSlotOfLocVars + 1
	} else {
		return 0
	}
}

/* code */

func (f *funcInfo) pc() int {
	return len(f.insts) - 1
}

func (f *funcInfo) fixSbx(pc, sBx int) {
	i := f.insts[pc]
	i = i << 18 >> 18                  // clear sBx
	i = i | uint32(sBx+MAXARG_sBx)<<14 // reset sBx
	f.insts[pc] = i
}

// todo: rename?
func (f *funcInfo) fixEndPC(name string, delta int) {
	for i := len(f.locVars) - 1; i >= 0; i-- {
		locVar := f.locVars[i]
		if locVar.name == name {
			locVar.endPC += delta
			return
		}
	}
}

func (f *funcInfo) emitABC(line, opcode, a, b, c int) {
	i := b<<23 | c<<14 | a<<6 | opcode
	f.insts = append(f.insts, uint32(i))
	f.lineNums = append(f.lineNums, uint32(line))
}

func (f *funcInfo) emitABx(line, opcode, a, bx int) {
	i := bx<<14 | a<<6 | opcode
	f.insts = append(f.insts, uint32(i))
	f.lineNums = append(f.lineNums, uint32(line))
}

func (f *funcInfo) emitAsBx(line, opcode, a, b int) {
	i := (b+MAXARG_sBx)<<14 | a<<6 | opcode
	f.insts = append(f.insts, uint32(i))
	f.lineNums = append(f.lineNums, uint32(line))
}

func (f *funcInfo) emitAx(line, opcode, ax int) {
	i := ax<<6 | opcode
	f.insts = append(f.insts, uint32(i))
	f.lineNums = append(f.lineNums, uint32(line))
}

// r[a] = r[b]
func (f *funcInfo) emitMove(line, a, b int) {
	f.emitABC(line, OP_MOVE, a, b, 0)
}

// r[a], r[a+1], ..., r[a+b] = nil
func (f *funcInfo) emitLoadNil(line, a, n int) {
	f.emitABC(line, OP_LOADNIL, a, n-1, 0)
}

// r[a] = (bool)b; if (c) pc++
func (f *funcInfo) emitLoadBool(line, a, b, c int) {
	f.emitABC(line, OP_LOADBOOL, a, b, c)
}

// r[a] = kst[bx]
func (f *funcInfo) emitLoadK(line, a int, k interface{}) {
	idx := f.indexOfConstant(k)
	if idx < (1 << 18) {
		f.emitABx(line, OP_LOADK, a, idx)
	} else {
		f.emitABx(line, OP_LOADKX, a, 0)
		f.emitAx(line, OP_EXTRAARG, idx)
	}
}

// r[a], r[a+1], ..., r[a+b-2] = vararg
func (f *funcInfo) emitVararg(line, a, n int) {
	f.emitABC(line, OP_VARARG, a, n+1, 0)
}

// r[a] = emitClosure(proto[bx])
func (f *funcInfo) emitClosure(line, a, bx int) {
	f.emitABx(line, OP_CLOSURE, a, bx)
}

// r[a] = {}
func (f *funcInfo) emitNewTable(line, a, nArr, nRec int) {
	f.emitABC(line, OP_NEWTABLE,
		a, Int2fb(nArr), Int2fb(nRec))
}

// r[a][(c-1)*FPF+i] := r[a+i], 1 <= i <= b
func (f *funcInfo) emitSetList(line, a, b, c int) {
	f.emitABC(line, OP_SETLIST, a, b, c)
}

// r[a] := r[b][rk(c)]
func (f *funcInfo) emitGetTable(line, a, b, c int) {
	f.emitABC(line, OP_GETTABLE, a, b, c)
}

// r[a][rk(b)] = rk(c)
func (f *funcInfo) emitSetTable(line, a, b, c int) {
	f.emitABC(line, OP_SETTABLE, a, b, c)
}

// r[a] = upval[b]
func (f *funcInfo) emitGetUpval(line, a, b int) {
	f.emitABC(line, OP_GETUPVAL, a, b, 0)
}

// upval[b] = r[a]
func (f *funcInfo) emitSetUpval(line, a, b int) {
	f.emitABC(line, OP_SETUPVAL, a, b, 0)
}

// r[a] = upval[b][rk(c)]
func (f *funcInfo) emitGetTabUp(line, a, b, c int) {
	f.emitABC(line, OP_GETTABUP, a, b, c)
}

// upval[a][rk(b)] = rk(c)
func (f *funcInfo) emitSetTabUp(line, a, b, c int) {
	f.emitABC(line, OP_SETTABUP, a, b, c)
}

// r[a], ..., r[a+c-2] = r[a](r[a+1], ..., r[a+b-1])
func (f *funcInfo) emitCall(line, a, nArgs, nRet int) {
	f.emitABC(line, OP_CALL, a, nArgs+1, nRet+1)
}

// return r[a](r[a+1], ... ,r[a+b-1])
func (f *funcInfo) emitTailCall(line, a, nArgs int) {
	f.emitABC(line, OP_TAILCALL, a, nArgs+1, 0)
}

// return r[a], ... ,r[a+b-2]
func (f *funcInfo) emitReturn(line, a, n int) {
	f.emitABC(line, OP_RETURN, a, n+1, 0)
}

// r[a+1] := r[b]; r[a] := r[b][rk(c)]
func (f *funcInfo) emitSelf(line, a, b, c int) {
	f.emitABC(line, OP_SELF, a, b, c)
}

// pc+=sBx; if (a) close all upvalues >= r[a - 1]
func (f *funcInfo) emitJmp(line, a, sBx int) int {
	f.emitAsBx(line, OP_JMP, a, sBx)
	return len(f.insts) - 1
}

// if not (r[a] <=> c) then pc++
func (f *funcInfo) emitTest(line, a, c int) {
	f.emitABC(line, OP_TEST, a, 0, c)
}

// if (r[b] <=> c) then r[a] := r[b] else pc++
func (f *funcInfo) emitTestSet(line, a, b, c int) {
	f.emitABC(line, OP_TESTSET, a, b, c)
}

func (f *funcInfo) emitForPrep(line, a, sBx int) int {
	f.emitAsBx(line, OP_FORPREP, a, sBx)
	return len(f.insts) - 1
}

func (f *funcInfo) emitForLoop(line, a, sBx int) int {
	f.emitAsBx(line, OP_FORLOOP, a, sBx)
	return len(f.insts) - 1
}

func (f *funcInfo) emitTForCall(line, a, c int) {
	f.emitABC(line, OP_TFORCALL, a, 0, c)
}

func (f *funcInfo) emitTForLoop(line, a, sBx int) {
	f.emitAsBx(line, OP_TFORLOOP, a, sBx)
}

// r[a] = op r[b]
func (f *funcInfo) emitUnaryOp(line, op, a, b int) {
	switch op {
	case TOKEN_OP_NOT:
		f.emitABC(line, OP_NOT, a, b, 0)
	case TOKEN_OP_BNOT:
		f.emitABC(line, OP_BNOT, a, b, 0)
	case TOKEN_OP_LEN:
		f.emitABC(line, OP_LEN, a, b, 0)
	case TOKEN_OP_UNM:
		f.emitABC(line, OP_UNM, a, b, 0)
	}
}

// r[a] = rk[b] op rk[c]
// arith & bitwise & relational
func (f *funcInfo) emitBinaryOp(line, op, a, b, c int) {
	if opcode, found := arithAndBitwiseBinops[op]; found {
		f.emitABC(line, opcode, a, b, c)
	} else {
		switch op {
		case TOKEN_OP_EQ:
			f.emitABC(line, OP_EQ, 1, b, c)
		case TOKEN_OP_NE:
			f.emitABC(line, OP_EQ, 0, b, c)
		case TOKEN_OP_LT:
			f.emitABC(line, OP_LT, 1, b, c)
		case TOKEN_OP_GT:
			f.emitABC(line, OP_LT, 1, c, b)
		case TOKEN_OP_LE:
			f.emitABC(line, OP_LE, 1, b, c)
		case TOKEN_OP_GE:
			f.emitABC(line, OP_LE, 1, c, b)
		}
		f.emitJmp(line, 0, 1)
		f.emitLoadBool(line, a, 0, 1)
		f.emitLoadBool(line, a, 1, 0)
	}
}
