package tui

import (
	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/huh"
)

type PodInfo struct {
	id               int
	name             string
	logs             string
	propsFilePath    string
	scenarioFilePath string
}

type ConfiguratorModel struct {
	currentView AppViewState
	currentPod  int
	pods        []PodInfo

	paginator         *paginator.Model
	filepicker        *FilePickerModule
	setupConfirmation *ConfirmationModel
	preparation       *PreparePodsModel
	configForm        *ConfigViewModel
	run               *TestRunModel
	err               error
}

type TestRunModel struct {
	runState     TestRunState
	namespace    string
	pods         []RunPodInfo
	isTableView  bool
	currentPod   int
	podViews     viewport.Model
	pages        paginator.Model
	confirm      *huh.Form
	table        string
	spinner      spinner.Model
	showSpinner  bool
	isConfirmed  bool
	showConfirm  bool
	prevRunState *TestRunState
}

type FilePickerModule struct {
	model        filepicker.Model
	selectedFile string
	mode         int
}

type RunConfigData struct {
	namespace  string
	podsAmount int
	pods       []RunPodInfo
}

type RunPodInfo struct {
	PodInfo

	runState   TestRunState
	err        error
	resultPath string
}

type AppViewState uint

const (
	Config AppViewState = iota
	FilePick
	PodsSetup
	ReviewSetup
	PreparePods
	Run
	Finish
)

type TestRunState uint

const (
	NotStarted TestRunState = iota
	StartConfirm
	InProgress
	Completed
	CancelConfirm
	Cancelled
	ResetConfirm
	Collected
)