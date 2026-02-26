package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	fynestorage "fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"

	appsettings "joyfuI/game-save-backup-manager/internal/appsettings"
	"joyfuI/game-save-backup-manager/internal/backup"
	"joyfuI/game-save-backup-manager/internal/model"
	"joyfuI/game-save-backup-manager/internal/pathutil"
	"joyfuI/game-save-backup-manager/internal/regutil"
	"joyfuI/game-save-backup-manager/internal/scan"
	"joyfuI/game-save-backup-manager/internal/storage"
)

const dbFile = "savelocation.db"
const backupDir = "backups"

type uiState struct {
	db                 *sql.DB
	window             fyne.Window
	manageWindow       fyne.Window
	allSaveScanResults []model.ScanResult
	foundSaveResults   []model.ScanResult
	mainListBox        *fyne.Container
}

type resizeAwareLayout struct {
	onResize func(size fyne.Size)
}

func (l *resizeAwareLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	for _, obj := range objects {
		obj.Move(fyne.NewPos(0, 0))
		obj.Resize(size)
	}
	if l.onResize != nil {
		l.onResize(size)
	}
}

func (l *resizeAwareLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	min := fyne.NewSize(0, 0)
	for _, obj := range objects {
		min = min.Max(obj.MinSize())
	}
	return min
}

func main() {
	if err := appsettings.EnsureInitialized(); err != nil {
		log.Fatalf("failed to initialize settings: %v", err)
	}

	db, err := storage.OpenAndInit(dbFile)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	state := &uiState{db: db}
	state.buildAndRun()
}

func (s *uiState) buildAndRun() {
	guiApp := app.New()
	window := guiApp.NewWindow("Game Save Backup Manager")
	window.Resize(fyne.NewSize(430, 520))
	s.window = window
	window.SetOnClosed(func() {
		if s.manageWindow != nil {
			s.manageWindow.Close()
		}
	})

	s.mainListBox = container.NewVBox()
	listScroll := container.NewVScroll(s.mainListBox)

	settingsButton := widget.NewButton("설정", func() {
		s.openSettingsDialog()
	})
	scanButton := widget.NewButton("스캔", func() {
		s.refreshScan()
	})

	topBar := container.NewHBox(scanButton, layout.NewSpacer(), settingsButton)
	title := widget.NewLabel("감지된 세이브 목록")
	content := container.NewBorder(topBar, nil, nil, nil, container.NewBorder(title, nil, nil, nil, listScroll))

	window.SetContent(content)

	s.refreshScan()
	window.ShowAndRun()
}

func (s *uiState) refreshScan() {
	locations, err := storage.LoadSaveLocations(s.db)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	s.allSaveScanResults = scan.SaveLocations(locations)
	s.foundSaveResults = make([]model.ScanResult, 0)
	for _, row := range s.allSaveScanResults {
		if row.Installed {
			s.foundSaveResults = append(s.foundSaveResults, row)
		}
	}

	s.renderFoundSaveList()
}

func (s *uiState) renderFoundSaveList() {
	s.mainListBox.Objects = nil

	if len(s.foundSaveResults) == 0 {
		s.mainListBox.Add(widget.NewLabel("감지된 세이브가 없습니다."))
		s.mainListBox.Refresh()
		return
	}

	for i, row := range s.foundSaveResults {
		loc := row.SaveLocation

		nameLabel := widget.NewLabel(loc.Name)
		nameLabel.Truncation = fyne.TextTruncateEllipsis

		backupButton := widget.NewButton("백업", func() {
			s.backupSaveLocation(loc)
		})
		openButton := widget.NewButton("경로 열기", func() {
			s.openSavePathInExplorer(loc)
		})

		buttons := container.NewHBox(backupButton, openButton)
		rowUI := container.NewBorder(nil, nil, nil, buttons, nameLabel)
		s.mainListBox.Add(rowUI)

		if i < len(s.foundSaveResults)-1 {
			s.mainListBox.Add(widget.NewSeparator())
		}
	}

	s.mainListBox.Refresh()
}

func (s *uiState) backupSaveLocation(loc model.SaveLocation) {
	switch strings.ToLower(strings.TrimSpace(loc.Type)) {
	case "zip":
		result, err := backup.CreateZipBackup(loc, backupDir)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		dialog.ShowInformation("백업", fmt.Sprintf("ZIP 백업 완료\n%s", result.ZipPath), s.window)
	case "reg":
		regPath, err := backup.CreateRegBackup(loc, backupDir)
		if err != nil {
			dialog.ShowError(err, s.window)
			return
		}
		dialog.ShowInformation("백업", fmt.Sprintf("REG 백업 완료\n%s", regPath), s.window)
	default:
		dialog.ShowInformation("백업", fmt.Sprintf("지원하지 않는 백업 타입입니다: %s", loc.Type), s.window)
	}
}

func (s *uiState) openSavePathInExplorer(loc model.SaveLocation) {
	if strings.EqualFold(strings.TrimSpace(loc.Type), "reg") {
		if err := regutil.OpenInEditor(loc.Path); err != nil {
			dialog.ShowError(err, s.window)
		}
		return
	}

	dir, err := resolveExplorerDirFromGlob(loc.Path)
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	if _, err := os.Stat(dir); err != nil {
		dialog.ShowError(fmt.Errorf("경로를 찾을 수 없습니다: %s", dir), s.window)
		return
	}

	if err := exec.Command("explorer", dir).Start(); err != nil {
		dialog.ShowError(fmt.Errorf("탐색기 실행 실패: %w", err), s.window)
	}
}

func resolveExplorerDirFromGlob(pathPattern string) (string, error) {
	resolvedPath, err := appsettings.ResolveSavePath(pathPattern)
	if err != nil {
		return "", err
	}

	expanded := strings.TrimSpace(resolvedPath)
	if expanded == "" {
		return "", fmt.Errorf("세이브 경로가 비어 있습니다")
	}

	if hasGlobWildcards(expanded) {
		return fixedPrefixDir(expanded), nil
	}

	info, err := os.Stat(expanded)
	if err == nil {
		if info.IsDir() {
			return expanded, nil
		}
		return filepath.Dir(expanded), nil
	}

	return filepath.Dir(expanded), nil
}

func hasGlobWildcards(path string) bool {
	return strings.ContainsAny(path, "*?[")
}

func fixedPrefixDir(pattern string) string {
	clean := filepath.Clean(pattern)
	volume := filepath.VolumeName(clean)
	rest := strings.TrimPrefix(clean, volume)
	rest = strings.TrimPrefix(rest, string(filepath.Separator))

	parts := strings.Split(rest, string(filepath.Separator))
	fixed := make([]string, 0, len(parts))
	for _, p := range parts {
		if strings.ContainsAny(p, "*?[") {
			break
		}
		fixed = append(fixed, p)
	}

	if len(fixed) == 0 {
		if volume != "" {
			return volume + string(filepath.Separator)
		}
		return string(filepath.Separator)
	}

	base := filepath.Join(fixed...)
	if volume != "" {
		return filepath.Join(volume+string(filepath.Separator), base)
	}
	return filepath.Join(string(filepath.Separator), base)
}

func (s *uiState) openManageWindow() {
	if s.manageWindow != nil {
		s.manageWindow.RequestFocus()
		return
	}

	manageWindow := fyne.CurrentApp().NewWindow("DB 관리")
	manageWindow.Resize(fyne.NewSize(1100, 700))
	s.manageWindow = manageWindow
	manageWindow.SetOnClosed(func() {
		s.manageWindow = nil
	})

	headers := []string{"게임", "타입", "백업 파일명", "세이브 경로", "매치 수"}
	selected := -1

	getRows := func() []model.ScanResult {
		return s.allSaveScanResults
	}

	manageTable := widget.NewTableWithHeaders(
		func() (int, int) {
			return len(getRows()), len(headers)
		},
		func() fyne.CanvasObject {
			label := widget.NewLabel("")
			label.Truncation = fyne.TextTruncateEllipsis
			return label
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			label := obj.(*widget.Label)
			row := getRows()[id.Row]
			switch id.Col {
			case 0:
				label.SetText(row.Name)
			case 1:
				label.SetText(row.Type)
			case 2:
				label.SetText(row.FileName)
			case 3:
				label.SetText(row.Path)
			case 4:
				label.SetText(fmt.Sprintf("%d", row.Matches))
			}
		},
	)
	manageTable.ShowHeaderColumn = false
	manageTable.CreateHeader = func() fyne.CanvasObject {
		label := widget.NewLabel("")
		label.Truncation = fyne.TextTruncateEllipsis
		return label
	}
	manageTable.UpdateHeader = func(id widget.TableCellID, obj fyne.CanvasObject) {
		if id.Row != -1 || id.Col < 0 || id.Col >= len(headers) {
			return
		}
		obj.(*widget.Label).SetText(headers[id.Col])
	}
	manageTable.OnSelected = func(id widget.TableCellID) {
		selected = id.Row
	}
	manageTable.OnUnselected = func(_ widget.TableCellID) {
		selected = -1
	}

	manageBaseWidths := []float32{170, 80, 180, 510, 100}
	manageTableLastWidth := float32(0)
	applyManageWidths := func(totalWidth float32) {
		if totalWidth <= 0 {
			return
		}
		if manageTableLastWidth > 0 && absf(totalWidth-manageTableLastWidth) < 1 {
			return
		}
		manageTableLastWidth = totalWidth
		applyColumnWidthsByRatio(manageTable, manageBaseWidths, totalWidth)
	}
	applyManageWidths(1100)

	refreshManage := func() {
		s.refreshScan()
		if selected >= len(getRows()) {
			selected = -1
		}
		manageTable.Refresh()
	}

	selectedManageLocation := func() (model.SaveLocation, bool) {
		rows := getRows()
		if selected < 0 || selected >= len(rows) {
			return model.SaveLocation{}, false
		}
		return rows[selected].SaveLocation, true
	}

	addButton := widget.NewButton("등록", func() {
		s.openUpsertDialog(manageWindow, nil, refreshManage)
	})
	editButton := widget.NewButton("수정", func() {
		loc, ok := selectedManageLocation()
		if !ok {
			dialog.ShowInformation("수정", "먼저 행을 선택하세요.", manageWindow)
			return
		}
		s.openUpsertDialog(manageWindow, &loc, refreshManage)
	})
	deleteButton := widget.NewButton("삭제", func() {
		loc, ok := selectedManageLocation()
		if !ok {
			dialog.ShowInformation("삭제", "먼저 행을 선택하세요.", manageWindow)
			return
		}
		s.openDeleteDialog(manageWindow, loc, refreshManage)
	})
	refreshButton := widget.NewButton("새로고침", refreshManage)

	topBar := container.NewHBox(addButton, editButton, deleteButton, refreshButton)
	manageTableWrap := container.New(&resizeAwareLayout{
		onResize: func(size fyne.Size) {
			applyManageWidths(size.Width)
		},
	}, manageTable)
	manageContent := container.NewBorder(topBar, nil, nil, nil, manageTableWrap)
	manageWindow.SetContent(manageContent)

	refreshManage()
	manageWindow.Show()
}

func (s *uiState) openSettingsDialog() {
	loaded, err := appsettings.Load()
	if err != nil {
		dialog.ShowError(err, s.window)
		return
	}

	ubisoftPathEntry := widget.NewEntry()
	ubisoftPathEntry.SetText(loaded.UbisoftConnectPath)
	ubisoftUserIDEntry := widget.NewEntry()
	ubisoftUserIDEntry.SetText(loaded.UbisoftConnectUserID)

	openFolderPicker := widget.NewButton("폴더 선택", func() {
		folderDialog := dialog.NewFolderOpen(func(list fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, s.window)
				return
			}
			if list == nil {
				return
			}
			chosen := normalizeLocalPathFromURI(list.Path())
			if strings.TrimSpace(chosen) == "" {
				return
			}
			ubisoftPathEntry.SetText(filepath.Clean(chosen))
		}, s.window)

		currentPath := strings.TrimSpace(pathutil.ExpandPathVariables(ubisoftPathEntry.Text))
		if currentPath != "" {
			if info, statErr := os.Stat(currentPath); statErr == nil && info.IsDir() {
				if uri := fynestorage.NewFileURI(currentPath); uri != nil {
					if listable, listErr := fynestorage.ListerForURI(uri); listErr == nil {
						folderDialog.SetLocation(listable)
					}
				}
			}
		}

		folderDialog.Show()
	})

	openManageButton := widget.NewButton("DB 관리", func() {
		s.openManageWindow()
	})
	pathRow := container.NewBorder(nil, nil, nil, openFolderPicker, ubisoftPathEntry)
	form := widget.NewForm(
		widget.NewFormItem("Ubisoft Connect USER ID", ubisoftUserIDEntry),
	)

	var settingsDialog dialog.Dialog

	saveButton := widget.NewButton("저장", func() {
		path := strings.TrimSpace(pathutil.ExpandPathVariables(ubisoftPathEntry.Text))
		if path == "" {
			dialog.ShowError(fmt.Errorf("Ubisoft Connect 설치 경로를 입력해 주세요"), s.window)
			return
		}

		toSave := appsettings.Settings{
			UbisoftConnectPath:   filepath.Clean(path),
			UbisoftConnectUserID: strings.TrimSpace(ubisoftUserIDEntry.Text),
		}
		if err := appsettings.Save(toSave); err != nil {
			dialog.ShowError(err, s.window)
			return
		}

		dialog.ShowInformation("설정", "설정을 저장했습니다.", s.window)
		if settingsDialog != nil {
			settingsDialog.Hide()
		}
	})
	saveButton.Importance = widget.HighImportance

	cancelButton := widget.NewButton("취소", func() {
		if settingsDialog != nil {
			settingsDialog.Hide()
		}
	})

	content := container.NewVBox(
		widget.NewLabel("Ubisoft Connect 설치 경로"),
		pathRow,
		form,
		openManageButton,
		container.NewGridWithColumns(2, cancelButton, saveButton),
	)

	settingsDialog = dialog.NewCustomWithoutButtons("설정", content, s.window)
	settingsDialog.Resize(fyne.NewSize(720, 280))
	settingsDialog.Show()
}

func normalizeLocalPathFromURI(uriPath string) string {
	decoded, err := url.PathUnescape(uriPath)
	if err != nil {
		decoded = uriPath
	}

	localPath := filepath.FromSlash(decoded)
	if runtime.GOOS == "windows" && len(localPath) >= 3 && localPath[0] == '\\' && localPath[2] == ':' {
		localPath = localPath[1:]
	}

	return localPath
}

func (s *uiState) openUpsertDialog(parent fyne.Window, existing *model.SaveLocation, onSaved func()) {
	nameEntry := widget.NewEntry()
	typeSelect := widget.NewSelect([]string{"zip", "reg"}, nil)
	fileNameEntry := widget.NewEntry()
	pathEntry := widget.NewEntry()

	typeSelect.SetSelected("zip")

	title := "세이브 경로 등록"
	if existing != nil {
		title = "세이브 경로 수정"
		nameEntry.SetText(existing.Name)
		switch strings.ToLower(strings.TrimSpace(existing.Type)) {
		case "reg":
			typeSelect.SetSelected("reg")
		default:
			typeSelect.SetSelected("zip")
		}
		fileNameEntry.SetText(existing.FileName)
		pathEntry.SetText(existing.Path)
	}

	items := []*widget.FormItem{
		widget.NewFormItem("게임 이름", nameEntry),
		widget.NewFormItem("타입", typeSelect),
		widget.NewFormItem("백업 파일명", fileNameEntry),
		widget.NewFormItem("세이브 경로", pathEntry),
	}
	formDialog := dialog.NewForm(title, "저장", "취소", items, func(ok bool) {
		if !ok {
			return
		}

		loc := model.SaveLocation{
			Name:     strings.TrimSpace(nameEntry.Text),
			Type:     strings.TrimSpace(typeSelect.Selected),
			FileName: strings.TrimSpace(fileNameEntry.Text),
			Path:     strings.TrimSpace(pathEntry.Text),
		}

		if err := validateSaveLocationInput(loc); err != nil {
			dialog.ShowError(err, parent)
			return
		}

		if existing == nil {
			if err := storage.InsertSaveLocation(s.db, loc); err != nil {
				dialog.ShowError(err, parent)
				return
			}
		} else {
			loc.RowID = existing.RowID
			if err := storage.UpdateSaveLocation(s.db, loc); err != nil {
				dialog.ShowError(err, parent)
				return
			}
		}

		onSaved()
	}, parent)
	formDialog.Resize(fyne.NewSize(640, 360))
	formDialog.Show()
}

func (s *uiState) openDeleteDialog(parent fyne.Window, loc model.SaveLocation, onDeleted func()) {
	confirm := dialog.NewConfirm(
		"세이브 경로 삭제",
		fmt.Sprintf("'%s' 항목을 삭제할까요?", loc.Name),
		func(ok bool) {
			if !ok {
				return
			}
			if err := storage.DeleteSaveLocation(s.db, loc.RowID); err != nil {
				dialog.ShowError(err, parent)
				return
			}
			onDeleted()
		},
		parent,
	)
	confirm.Show()
}

func validateSaveLocationInput(loc model.SaveLocation) error {
	if loc.Name == "" {
		return fmt.Errorf("게임 이름은 필수입니다")
	}
	if loc.Type == "" {
		return fmt.Errorf("타입은 필수입니다")
	}
	if loc.FileName == "" {
		return fmt.Errorf("백업 파일명은 필수입니다")
	}
	if strings.Contains(loc.FileName, ".") {
		return fmt.Errorf("백업 파일명에 확장자를 포함하면 안 됩니다")
	}
	if loc.Path == "" {
		return fmt.Errorf("세이브 경로는 필수입니다")
	}

	switch strings.ToLower(strings.TrimSpace(loc.Type)) {
	case "zip":
		patternForValidation := substituteKnownSavePathTokensForValidation(loc.Path)
		if _, err := filepath.Match(patternForValidation, "sample"); err != nil {
			return fmt.Errorf("Glob 패턴이 잘못되었습니다: %w", err)
		}
	case "reg":
		if err := regutil.ValidatePath(loc.Path); err != nil {
			return err
		}
	default:
		return fmt.Errorf("지원하지 않는 타입입니다: %s", loc.Type)
	}

	return nil
}

func substituteKnownSavePathTokensForValidation(path string) string {
	resolved := replaceTokenInsensitive(path, "{{ubisoftconnect-path}}", `C:\Ubisoft\Ubisoft Game Launcher`)
	return replaceTokenInsensitive(resolved, "{{ubisoftconnect-userid}}", "user-id")
}

func replaceTokenInsensitive(input, token, replacement string) string {
	lowerInput := strings.ToLower(input)
	lowerToken := strings.ToLower(token)

	if !strings.Contains(lowerInput, lowerToken) {
		return input
	}

	var builder strings.Builder
	for {
		idx := strings.Index(lowerInput, lowerToken)
		if idx < 0 {
			builder.WriteString(input)
			break
		}

		builder.WriteString(input[:idx])
		builder.WriteString(replacement)
		input = input[idx+len(token):]
		lowerInput = lowerInput[idx+len(token):]
	}

	return builder.String()
}

func applyColumnWidthsByRatio(table *widget.Table, baseWidths []float32, totalWidth float32) {
	if len(baseWidths) == 0 {
		return
	}

	usableWidth := totalWidth - 16
	if usableWidth < float32(len(baseWidths))*60 {
		usableWidth = float32(len(baseWidths)) * 60
	}

	baseSum := float32(0)
	for _, w := range baseWidths {
		baseSum += w
	}
	if baseSum <= 0 {
		return
	}

	scale := usableWidth / baseSum
	for i, w := range baseWidths {
		table.SetColumnWidth(i, w*scale)
	}
}

func absf(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
