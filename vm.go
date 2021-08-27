package main

/*
#cgo LDFLAGS: -ldl

#include <stdlib.h>
#include <dlfcn.h>

typedef struct eniac_t {
	unsigned long long cycles;
	int error_code;
	int rollback;
	char acc[20][12];
	int ft[3][104][14];
} ENIAC;

typedef void* (*vmNewFunc) (void);
typedef int (*vmImportFunc) (void *vm, ENIAC* eniac);
typedef void (*vmStepFunc) (void *vm);
typedef void (*vmStepToFunc) (void *vm, unsigned long long cycle);
typedef void (*vmExportFunc) (void *vm, ENIAC* eniac);
typedef void (*vmFreeFunc) (void *vm);

void* bridge_vm_new(vmNewFunc f) {
	return f();
}
int bridge_vm_import(vmImportFunc f, void *vm, ENIAC* eniac) {
	return f(vm, eniac);
}
void bridge_vm_step(vmStepFunc f, void *vm) {
	f(vm);
}
void bridge_vm_step_to(vmStepToFunc f, void *vm, unsigned long long cycle) {
	f(vm, cycle);
}
void bridge_vm_export(vmExportFunc f, void *vm, ENIAC* eniac) {
	f(vm, eniac);
}
void bridge_vm_free(vmFreeFunc f, void *vm) {
	f(vm);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// VM wraps a shared library that imports a checkpoint of ENIAC state, steps a
// higher level simulation, and exports another checkpoint.  This permits
// cross-validation, as well as mixing faster instruction level simulation with
// eniacsim's pulse level simulation.
type VM struct {
	lib *sharedLibrary // shared library handles
	vm  unsafe.Pointer // VM's state

	validState bool    // true if eniac state is up-to-date
	eniac      C.ENIAC // eniacsim's most recent exported state
	nextEniac  C.ENIAC // vm's updated eniac state

	ErrorCode int  // if nonzero, VM error
	ftDumped  bool // only dump ft data once
}

type sharedLibrary struct {
	handle   unsafe.Pointer
	vmNew    C.vmNewFunc
	vmFree   C.vmFreeFunc
	vmImport C.vmImportFunc
	vmExport C.vmExportFunc
	vmStep   C.vmStepFunc
	vmStepTo C.vmStepToFunc
}

// Returns a new VM wrapper.  path points to a shared library for an
// instruction level simulator
func NewVM(path string) *VM {
	if len(path) == 0 {
		return &VM{}
	}
	lib, err := loadLibrary(path)
	if err != nil {
		panic(err)
	}
	vm := C.bridge_vm_new(lib.vmNew)
	return &VM{lib: lib, vm: vm}
}

// Steps the VM forward and verifies its state against the latest ENIAC state checkpoint
func (vm *VM) StepAndVerify() {
	if vm.lib == nil {
		return
	}
	if !vm.validState {
		// Setup VM state to match the current ENIAC state.
		vm.exportEniacState()
		if C.bridge_vm_import(vm.lib.vmImport, vm.vm, &vm.eniac) != 0 {
			panic("vm: error importing eniac state")
		}
		vm.validState = true
		return
	}

	// The eniacsim simulation should now be one checkpoint ahead of the VM
	// simulation. Run the VM and make sure that its output matches.
	C.bridge_vm_step(vm.lib.vmStep, vm.vm)
	C.bridge_vm_export(vm.lib.vmExport, vm.vm, &vm.nextEniac)
	if vm.nextEniac.rollback != 0 {
		// If I/O happened, resync at next checkpoint
		vm.validState = false
		return
	}
	vm.exportEniacState()
	if err := vm.compareEniacState(); err != nil {
		panic(err)
	}
}

// Steps the VM ahead of eniacsim up to but not exceeding cycle, and re-imports
// its state.  Returns early if there is I/O or break/halt.
func (vm *VM) StepAhead(cycle int64) {
	if vm.lib == nil {
		return
	}
	vm.exportEniacState()
	if C.bridge_vm_import(vm.lib.vmImport, vm.vm, &vm.eniac) != 0 {
		panic("vm: error importing eniac state")
	}
	C.bridge_vm_step_to(vm.lib.vmStepTo, vm.vm, C.ulonglong(cycle))
	C.bridge_vm_export(vm.lib.vmExport, vm.vm, &vm.eniac)
	vm.importEniacState()
	// In case we switch to checking, resync to a new checkpoint
	vm.validState = false
}

// Exports a subset of eniac state to struct ENIAC which VM can read.
func (vm *VM) exportEniacState() {
	vm.eniac.cycles = C.ulonglong(cycle.AddCycle)
	vm.eniac.error_code = 0
	vm.eniac.rollback = 0
	for i := 0; i < 20; i++ {
		value := u.Accumulator[i].Value()
		vm.eniac.acc[i][0] = C.char(value[0])
		for j := 0; j < 10; j++ {
			vm.eniac.acc[i][1+j] = C.char(value[2+j])
		}
		vm.eniac.acc[i][11] = 0
	}
	if !vm.ftDumped {
		for t := 0; t < 3; t++ {
			for r := 0; r < 104; r++ {
				for d := 0; d < 14; d++ {
					vm.eniac.ft[t][r][d] = C.int(u.Ft[t].GetDigit(r, d))
				}
			}
		}
		vm.ftDumped = true
	}
}

// Imports struct ENIAC state, as written by VM.
func (vm *VM) importEniacState() {
	vm.ErrorCode = int(vm.eniac.error_code)
	if vm.ErrorCode != 0 {
		panic(fmt.Errorf("vm error state: %d\n", vm.ErrorCode))
	}
	cycle.AddCycle = int64(vm.eniac.cycles)
	for i := 0; i < 20; i++ {
		if vm.eniac.acc[i][0] == 0 {
			// Skip don't-care accumulators which the vm doesn't model
			continue
		}
		u.Accumulator[i].SetValue(C.GoBytes(unsafe.Pointer(&vm.eniac.acc[i]), 11))
	}
}

// Checks that eniacsim and VM's eniac state match at a checkpoint.
func (vm *VM) compareEniacState() error {
	if vm.eniac.cycles != vm.nextEniac.cycles {
		return fmt.Errorf("vm cycles mismatch: E:%v V:%v\n", vm.eniac.cycles, vm.nextEniac.cycles)
	}
	if vm.nextEniac.error_code != 0 {
		return fmt.Errorf("vm error: %v\n", vm.nextEniac.error_code)
	}
	if vm.nextEniac.rollback != 0 {
		return fmt.Errorf("vm signaled rollback\n")
	}
	for i := 0; i < 20; i++ {
		for j := 0; j < 11; j++ {
			vmValue := vm.nextEniac.acc[i][j]
			// nul signals don't care - skip comparison of IR and EX to avoid having
			// to model fetch uarch in VM simulator
			if vmValue != 0 && vmValue != vm.eniac.acc[i][j] {
				e := C.GoStringN(&vm.eniac.acc[i][0], 11)
				v := C.GoStringN(&vm.nextEniac.acc[i][0], 11)
				return fmt.Errorf("vm acc %d mismatch: E:%s V:%s\n", i, e, v)
			}
		}
	}
	return nil
}

// Closes shared library and frees resources.
func (vm *VM) Close() {
	if vm.lib == nil {
		return
	}
	if vm.vm != nil {
		C.bridge_vm_free(vm.lib.vmFree, vm.vm)
	}
	if vm.lib.handle != nil {
		C.dlclose(vm.lib.handle)
		vm.lib = nil
	}
}

func loadLibrary(path string) (*sharedLibrary, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	handle := C.dlopen(cPath, C.RTLD_LAZY)
	if handle == nil {
		return nil, fmt.Errorf("error opening %s", path)
	}

	vmNew := C.vmNewFunc(dlsym("vm_new", handle))
	vmImport := C.vmImportFunc(dlsym("vm_import", handle))
	vmStep := C.vmStepFunc(dlsym("vm_step", handle))
	vmStepTo := C.vmStepToFunc(dlsym("vm_step_to", handle))
	vmExport := C.vmExportFunc(dlsym("vm_export", handle))
	vmFree := C.vmFreeFunc(dlsym("vm_free", handle))
	if vmNew == nil || vmImport == nil || vmStep == nil || vmExport == nil || vmFree == nil {
		return nil, fmt.Errorf("missing symbol(s) in %s", path)
	}
	return &sharedLibrary{
		handle:   handle,
		vmNew:    vmNew,
		vmImport: vmImport,
		vmStep:   vmStep,
		vmStepTo: vmStepTo,
		vmExport: vmExport,
		vmFree:   vmFree,
	}, nil
}

func dlsym(name string, handle unsafe.Pointer) unsafe.Pointer {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	return C.dlsym(handle, cName)
}
