# 게임 세이브 백업 매니저

이 프로젝트는 **Codex로만** 만들어졌습니다.

Windows에서 게임 세이브를 `zip` 또는 `reg` 형식으로 백업하는 GUI 도구입니다.

## 주요 기능
- 감지된 세이브 목록 표시(설치/존재 확인된 항목만)
- `DB 관리` 창에서 등록/수정/삭제
- `zip` 백업: 파일/폴더 기반 세이브를 ZIP으로 백업
- `reg` 백업: 레지스트리 키/값을 `.reg` 파일로 백업

## 백업 타입별 경로 규칙
- `zip`
  - `path`는 파일 경로 또는 Glob 패턴
  - `%APPDATA%` 같은 환경변수 사용 가능
- `reg`
  - 경로가 `\\`로 끝나면: **키 전체 백업**
  - 경로가 `\\`로 안 끝나면: **값 단일 백업**(마지막 세그먼트가 값 이름)

예시:
- 키 전체: `HKCU\\Software\\MyGame\\`
- 값 단일: `HKEY_LOCAL_MACHINE\\SOFTWARE\\Disney Interactive\\Hercules\\1.00\\Config`

## DB 스키마 (`savelocation`)
- `name`: 게임 이름
- `type`: `zip` 또는 `reg`
- `filename`: 백업 파일명(확장자 제외)
- `path`: 세이브 경로

## 실행
```powershell
go run .
```

## 빌드
```powershell
go build ./...
fyne package --target windows --source-dir . --icon assets\app-icon.png --name game-save-backup-manager
New-Item -ItemType Directory -Force out | Out-Null
Move-Item -Force game-save-backup-manager.exe out\game-save-backup-manager.exe
```

## 산출물/아이콘
- 실행 파일: `out\game-save-backup-manager.exe`
- 아이콘: `assets/app-icon.png`, `assets/app-icon.ico`
