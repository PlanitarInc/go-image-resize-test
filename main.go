package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v1"
)

const (
	TMP_DIR = "./tmp"
)

func main() {
	cmdList := &cobra.Command{
		Use:   "list",
		Short: "List the available resizers",
		Run:   cmdListDo,
	}

	cmdRun := &cobra.Command{
		Use: "run [OPTIONS] <image1> [<image2> [<image3> [...]]]",
		Run: cmdRunDo,
	}

	cmdBench := &cobra.Command{
		Use:   "bench [OPTIONS]",
		Short: "Clean the tmp directory",
		Run:   cmdBenchDo,
	}

	cmdClean := &cobra.Command{
		Use:   "clean",
		Short: "Clean the tmp directory",
		Run:   cmdCleanDo,
	}

	rootCmd := &cobra.Command{
		Use: "resize",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if useNativeJpeg {
				fmt.Printf("[I] Using GO native JPEG\n")
				jpegBackend = &JpegNative{}
			} else {
				fmt.Printf("[I] Using JPEG turbo\n")
				jpegBackend = &JpegTurbo{}
			}
		},
	}

	rootCmd.PersistentFlags().StringVarP(&configFile,
		"config-file", "c", "./config.yaml", "config file, e.g. ./config.yaml")
	rootCmd.PersistentFlags().StringVarP(&destPath,
		"out", "o", TMP_DIR, "output dir name")
	rootCmd.PersistentFlags().IntVarP(&dstWidth,
		"width", "W", 0, "width of the target image")
	rootCmd.PersistentFlags().IntVarP(&dstHeight,
		"height", "H", 0, "height of the target image")
	rootCmd.PersistentFlags().BoolVar(&useNativeJpeg,
		"native", false, "use native jpeg")
	rootCmd.AddCommand(cmdList, cmdRun, cmdBench, cmdClean)

	rootCmd.Execute()
}

var errlog = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)

var (
	configFile    = "./config.yaml"
	destPath      = TMP_DIR
	dstWidth      = 0
	dstHeight     = 0
	useNativeJpeg = false

	jpegBackend JpegBE
	allResizers ResizerList
)

func RegisterResizer(resizers ...ResizerDesc) {
	allResizers = append(allResizers, resizers...)
}

func getResizers() (ResizerList, error) {
	if configFile == "" {
		return allResizers, nil
	}

	cfg := struct {
		ResizerNames []string `yaml:"resizers"`
	}{}

	bs, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file: %s", err)
	}

	err = yaml.Unmarshal(bs, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse config file: %s", err)
	}

	list := make([]ResizerDesc, len(cfg.ResizerNames))
	for i := range cfg.ResizerNames {
		if r := allResizers.Find(cfg.ResizerNames[i]); r == nil {
			return nil, fmt.Errorf("Unknown resizer: %s", cfg.ResizerNames[i])
		} else {
			list[i] = *r
		}
	}
	return list, nil
}

func cmdListDo(cmd *cobra.Command, args []string) {
	fmt.Printf("resizers:\n")
	// Assume the resizers are sorted by library
	lastLib := ""
	for i := range allResizers {
		if lastLib != allResizers[i].Library {
			lastLib = allResizers[i].Library
			if i > 0 {
				fmt.Println("")
			}
			fmt.Printf(" # %s (%s)\n", allResizers[i].Library, allResizers[i].Url)
		}
		fmt.Printf(" - %s\n", allResizers[i].Name)
	}
}

func cmdRunDo(cmd *cobra.Command, args []string) {
	fmt.Printf("bench: %v\n", args)

	resizers, err := getResizers()
	if err != nil {
		errlog.Fatalf("%s", err)
	}

	if dstWidth < 0 || dstHeight < 0 {
		errlog.Fatalf("Negative target dimensions: %d %d",
			dstWidth, dstHeight)
	}

	if dstWidth == 0 && dstHeight == 0 {
		errlog.Fatalf("Empty target dimensions: %d %d", dstWidth, dstHeight)
	}

	if len(args) == 0 {
		errlog.Fatalf("No images provided")
	}

	if err := exec.Command("mkdir", "-p", destPath).Run(); err != nil {
		errlog.Fatalf("Failed to create output dir: %s", err)
	}

	srcFile := args[0]
	if len(args) > 1 {
		errlog.Printf("Taking only the first argument '%s', others are ignored\n",
			srcFile)
	}

	srcImage, err := os.Open(srcFile)
	if err != nil {
		errlog.Fatalf("Failed to open '%s': %s", srcFile, err)
	}
	defer srcImage.Close()

	origSize := seekerLen(srcImage)
	allStats := ResizeStatsList{}

	// Fill the caches
	io.Copy(ioutil.Discard, srcImage)

	for _, resizer := range resizers {
		srcImage.Seek(0, 0)

		dirname := destPath
		if dirname == "" {
			dirname = path.Dir(srcFile)
		}
		srcBasename := path.Base(srcFile)
		ext := path.Ext(srcBasename)
		filename := strings.TrimSuffix(srcBasename, ext)
		dstFile := fmt.Sprintf("%s/%s.%s.%dx%d%s",
			dirname, filename, resizer.Name, dstWidth, dstHeight, ext)

		dstImage, err := os.Create(dstFile)
		if err != nil {
			errlog.Fatalf("%s: Failed to open '%s': %s", resizer.Name, dstFile, err)
		}
		defer dstImage.Close()

		stats := ResizerStats{Resizer: resizer}
		stats.Start()

		if err := resizer.Instance.Resize(dstImage, srcImage, dstWidth, dstHeight); err != nil {
			errlog.Fatalf("%s: Failed to resize '%s': %s", resizer.Name, srcFile, err)
		}

		stats.Stop()
		if s, err := dstImage.Stat(); err != nil {
			errlog.Fatalf("%s: Failed to stat '%s': %s", resizer.Name, dstFile, err)
		} else {
			stats.TotalSize = s.Size()
		}

		allStats = append(allStats, stats)
	}

	//      sort.Sort(ByAvgDuration{allStats})

	w := os.Stdout
	fmt.Fprintf(w, "Original image: name=%s, size=%s\n\n", srcFile, b2s(origSize))
	formatHeader := "|%-32s|%-18s|%-18s|%-18s|%-18s|\n"
	formatRow := "| %-30s | %-16s | %-16s | %-16s | %-16d |\n"
	fmt.Fprintf(w, formatHeader, " Name", " Time", " Size", " Bytes", "Mallocs")
	s18 := strings.Repeat("-", 17) + ":"
	fmt.Fprintf(w, formatHeader, strings.Repeat("-", 32), s18, s18, s18, s18)
	for _, s := range allStats {
		fmt.Fprintf(w, formatRow,
			s.Resizer.Name,
			s.TotalDuration,
			b2s(s.TotalSize)+" ("+percent(s.TotalSize, origSize)+")",
			b2s(s.TotalBytes),
			s.TotalMallocs)
	}
	fmt.Fprintln(w)
}

func cmdBenchDo(cmd *cobra.Command, args []string) {
	fmt.Printf("bench: %v\n", args)

	resizers, err := getResizers()
	if err != nil {
		errlog.Fatalf("%s", err)
	}

	if dstWidth < 0 || dstHeight < 0 {
		errlog.Fatalf("Negative target dimensions: %d %d",
			dstWidth, dstHeight)
	}

	if dstWidth == 0 && dstHeight == 0 {
		errlog.Fatalf("Empty target dimensions: %d %d", dstWidth, dstHeight)
	}

	if len(args) == 0 {
		errlog.Fatalf("No images provided")
	}

	srcFile := args[0]
	if len(args) > 1 {
		errlog.Printf("Taking only the first argument '%s', others are ignored\n",
			srcFile)
	}

	srcBytes, err := ioutil.ReadFile(srcFile)
	if err != nil {
		errlog.Fatalf("Failed to read '%s': %s", srcFile, err)
	}

	srcImage := bytes.NewReader(srcBytes)
	origSize := int64(len(srcBytes))
	allStats := ResizeStatsList{}

	N := 100

	for _, resizer := range resizers {
		stats := ResizerStats{Resizer: resizer, NRuns: int64(N)}

		stats.Start()
		for i := 0; i < N; i++ {
			srcImage.Seek(0, 0)
			countWriter := DiscardCount{0}

			err := resizer.Instance.Resize(&countWriter, srcImage, dstWidth, dstHeight)
			if err != nil {
				errlog.Printf("%s: Failed to resize '%s': %s", resizer.Name, srcFile, err)
				break
			}

			stats.TotalSize += countWriter.N
		}
		stats.Stop()

		allStats = append(allStats, stats)
	}

	//      sort.Sort(ByAvgDuration{allStats})

	w := os.Stdout
	fmt.Fprintf(w, "Original image: name=%s, size=%s\n\n", srcFile, b2s(origSize))
	fmt.Fprintf(w, "#iterations: %d\n", N)
	formatHeader := "|%-32s|%-18s|%-18s|%-18s|%-18s|\n"
	formatRow := "| %-30s | %-16s | %-16s | %-16s | %-16d |\n"
	fmt.Fprintf(w, formatHeader, " Name", " Avg Time", " Size", " Avg Bytes", " Avg Mallocs")
	s18 := strings.Repeat("-", 17) + ":"
	fmt.Fprintf(w, formatHeader, strings.Repeat("-", 32), s18, s18, s18, s18)
	for _, s := range allStats {
		fmt.Fprintf(w, formatRow,
			s.Resizer.Name,
			s.AvgDuration,
			b2s(s.AvgSize),
			b2s(s.AvgBytes),
			s.AvgMalloc)
	}
	fmt.Fprintln(w)
}

func cmdCleanDo(cmd *cobra.Command, args []string) {
	rmCmd := exec.Command("rm", "-rf", fmt.Sprintf(TMP_DIR))

	stderr := bytes.Buffer{}
	rmCmd.Stderr = &stderr

	err := rmCmd.Run()
	if err != nil {
		errlog.Fatalf(stderr.String())
	}
}

type ResizeStatsList []ResizerStats

func (s ResizeStatsList) Len() int      { return len(s) }
func (s ResizeStatsList) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByAvgDuration struct{ ResizeStatsList }

func (b ByAvgDuration) Less(i, j int) bool {
	return b.ResizeStatsList[i].AvgDuration < b.ResizeStatsList[j].AvgDuration
}

type ResizerStats struct {
	Resizer ResizerDesc

	NRuns         int64
	TotalSize     int64
	TotalDuration time.Duration
	TotalBytes    int64
	TotalMallocs  int64
	AvgSize       int64
	AvgDuration   time.Duration
	AvgBytes      int64
	AvgMalloc     int64

	startTime, endTime time.Time
	startMem, endMem   runtime.MemStats
}

func (s *ResizerStats) Start() {
	s.startTime = time.Now()
	runtime.ReadMemStats(&s.startMem)
}

func (s *ResizerStats) Stop() {
	s.endTime = time.Now()
	runtime.ReadMemStats(&s.endMem)

	s.TotalDuration = s.endTime.Sub(s.startTime)
	s.TotalBytes = int64(s.endMem.TotalAlloc - s.startMem.TotalAlloc)
	s.TotalMallocs = int64(s.endMem.Mallocs - s.startMem.Mallocs)
	if s.NRuns > 0 {
		s.AvgDuration = time.Duration(int64(s.TotalDuration) / s.NRuns)
		s.AvgBytes = s.TotalBytes / s.NRuns
		s.AvgMalloc = s.TotalMallocs / s.NRuns
		s.AvgSize = s.TotalSize / s.NRuns
	}
}
