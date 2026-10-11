package mobile

import (
	"fmt"
	"io"
	"os"
	"runtime"

	"kylix/pkg/llvmgen"
)

// BuildAndroid cross-compiles the flashcard core (not server.klx) into
// jniLibs and checks the ELF exports. abiName is amd64 or arm64.
func BuildAndroid(root, abiName string, w io.Writer) error {
	abi, err := parseAndroidABI(abiName)
	if err != nil {
		return err
	}
	root, err = needRoot(root)
	if err != nil {
		return err
	}
	if w == nil {
		w = os.Stderr
	}
	probe := llvmgen.ProbeAndroidNdk()
	if probe.Found == nil {
		return probe.MissError()
	}
	so := androidSO(root, abi)
	fmt.Fprintf(w, "[vocab-android] %s → %s\n", abi.Target, so)
	// IntermediatesDir left empty: compileLLVM keeps .ll/.o in a temp dir
	// it deletes, not beside the .klx sources.
	opts := llvmgen.CompileOpts{Shared: true, Target: abi.Target}
	if err := compileLLVM(vocabCoreSources(root), so, opts); err != nil {
		return err
	}
	if err := verifyAndroid(root, so, abi, w); err != nil {
		return err
	}
	fmt.Fprintf(w, "[vocab-android] wrote %s\n", so)
	fmt.Fprintln(w, EmulatorSentence)
	fmt.Fprintf(w, "android artifact OK (%s)\n", abi.Name)
	return nil
}

// ArtifactAndroid checks an existing flashcard .so.
func ArtifactAndroid(root, so, abiName string, w io.Writer) error {
	abi, err := parseAndroidABI(abiName)
	if err != nil {
		return err
	}
	root, err = needRoot(root)
	if err != nil {
		return err
	}
	if w == nil {
		w = os.Stderr
	}
	if err := verifyAndroid(root, so, abi, w); err != nil {
		return err
	}
	fmt.Fprintln(w, EmulatorSentence)
	fmt.Fprintf(w, "android artifact OK (%s)\n", abi.Name)
	return nil
}

// BuildIOS links libkylixvocab.a. On any host other than macOS it returns
// IosHostError and does not write an archive.
func BuildIOS(root, kindName string, w io.Writer) error {
	kind, err := parseIosKind(kindName)
	if err != nil {
		return err
	}
	if runtime.GOOS != "darwin" {
		return llvmgen.IosHostError()
	}
	root, err = needRoot(root)
	if err != nil {
		return err
	}
	if w == nil {
		w = os.Stderr
	}
	archive := iosArchive(root)
	fmt.Fprintf(w, "[vocab-ios] %s → %s\n", kind.Target, archive)
	opts := llvmgen.CompileOpts{Shared: true, Target: kind.Target}
	if err := compileLLVM(vocabCoreSources(root), archive, opts); err != nil {
		return err
	}
	if err := verifyIOS(root, archive, kind, w); err != nil {
		return err
	}
	fmt.Fprintf(w, "[vocab-ios] wrote %s\n", archive)
	fmt.Fprintf(w, "ios artifact OK (%s)\n", kind.Name)
	return nil
}

// ArtifactIOS checks an existing flashcard static library.
func ArtifactIOS(root, archive, kindName string, w io.Writer) error {
	kind, err := parseIosKind(kindName)
	if err != nil {
		return err
	}
	if runtime.GOOS != "darwin" {
		return llvmgen.IosHostError()
	}
	root, err = needRoot(root)
	if err != nil {
		return err
	}
	if w == nil {
		w = os.Stderr
	}
	if err := verifyIOS(root, archive, kind, w); err != nil {
		return err
	}
	fmt.Fprintf(w, "ios artifact OK (%s)\n", kind.Name)
	return nil
}

func runtimeDarwin() bool { return runtime.GOOS == "darwin" }
