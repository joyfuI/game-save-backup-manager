# 게임 세이브 백업 매니저

이 프로젝트는 **Codex로만** 만들어졌습니다.

Windows에서 게임 세이브를 `zip` 또는 `reg` 형식으로 백업하는 GUI 도구입니다.

## 주요 기능
- 감지된 세이브 목록 표시(실제로 세이브 경로가 매치되는 항목만)
- 각 항목별 `백업`, `경로 열기`
- `DB 관리` 창에서 등록/수정/삭제
- `zip` 백업: 파일/폴더 세이브를 ZIP으로 백업
- `reg` 백업: 레지스트리 키/값을 `.reg` 파일로 백업

## 설정
설정 다이얼로그에서 아래 항목을 관리합니다.
- `Steam 설치 경로`
- `Steam USER ID`
- `Ubisoft Connect 설치 경로`
- `Ubisoft Connect USER ID`

설정은 실행 파일과 같은 경로의 ini 파일에 저장됩니다.
- 파일명: `<exe basename>.ini` (예: `game-save-backup-manager.ini`)
- 키:
  - `steam_path`
  - `steam_userid`
  - `ubisoft_connect_path`
  - `ubisoft_connect_userid`

## 세이브 경로 플레이스홀더
`zip` 타입 path에서 아래 플레이스홀더를 사용할 수 있습니다.
- `{{steam-path}}` => 설정의 Steam 설치 경로
- `{{steam-userid}}` => 설정의 Steam USER ID
- `{{ubisoftconnect-path}}` => 설정의 Ubisoft Connect 설치 경로
- `{{ubisoftconnect-userid}}` => 설정의 Ubisoft Connect USER ID

예시:
- DB path: `{{ubisoftconnect-path}}\savegames\{{ubisoftconnect-userid}}\4\*`
- 실제 탐색: `C:\Program Files (x86)\Ubisoft\Ubisoft Game Launcher\savegames\12345\4\*`

주의:
- 스캔/실제 파일 접근은 치환된 실제 경로로 수행됩니다.
- ZIP 내부 경로는 플레이스홀더 원문을 유지합니다.
- 필요한 설정값이 비어 있으면 해당 항목은 감지/백업이 실패할 수 있습니다.

## reg 경로 규칙
- path가 `\\`로 끝나면: 키 전체 백업
- path가 `\\`로 끝나지 않으면: 마지막 세그먼트를 값 이름으로 간주해 값 단일 백업

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
fyne package -os windows
# 배포용
fyne package -os windows -release
```

## 산출물/아이콘
- 실행 파일: `game-save-backup-manager.exe` (프로젝트 루트)
- 아이콘: `app-icon.png`
