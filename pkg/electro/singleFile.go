package electro

import (
	"app/pkg/simplepack"
	"fmt"
)

func publishSingleFile(pathOutputDir string) error {
	err := packSite(pathOutputDir)
	if err != nil {
		return fmt.Errorf("error packing site into single file: %w", err)
	}

	// FIXME: implement coping of single file to requested output_single_file path
	return nil
}

func packSite(pathOutputDir string) error {
	var err error

	qlog.InfoPrint("Packing site...")

	pathFile := pathOutputDir + "/index.raw.html"
	pathFileStage1 := pathOutputDir + "/index.packed.stage1.html"
	pathFileStage2 := pathOutputDir + "/index.packed.stage2.html"
	pathFileStage3 := pathOutputDir + "/index.packed.stage3.html"
	pathFilePacked := pathOutputDir + "/index.html"

	// ------------------
	// STAGE 1: Pack
	// ------------------
	qlog.InfoPrintf("Packing %q to %q...", pathFile, pathFilePacked)
	enableMinify := true
	err = simplepack.Pack(pathFile, pathFileStage1, enableMinify)
	if err != nil {
		return fmt.Errorf("error packing site (stage 1): %w", err)
	}

	// ------------------
	// STAGE 2: Inline Images
	// ------------------
	qlog.InfoPrintf("Inlining images to  %q...", pathFileStage2)
	err = makeHTMLImagesInline(pathFileStage1, pathFileStage2)
	if err != nil {
		return fmt.Errorf("error inlining images: %w", err)
	}

	// ------------------
	// STAGE 3: Inline HTML Fonts
	// ------------------
	qlog.InfoPrintf("Inlining fonts to  %q...", pathFileStage3)
	err = makeHTMLFontsInline(pathFileStage2, pathFileStage3)
	if err != nil {
		return fmt.Errorf("error inlining fonts: %w", err)
	}

	// ------------------
	// STAGE 3: Inline HTML Icons
	// ------------------
	qlog.InfoPrintf("Inlining icons to  %q...", pathFilePacked)
	err = makeHTMLIconsInline(pathFileStage3, pathFilePacked)
	if err != nil {
		return fmt.Errorf("error inlining icons: %w", err)
	}

	return nil
}
