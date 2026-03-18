# AGENTS.md

## 목적
새 세션 에이전트가 코드 전체를 다시 파악하지 않고도, 변경 시 깨지기 쉬운 운영 규칙과 합의사항만 빠르게 이어받도록 한다.

## 절대 규칙
- 실행 파일 산출물은 프로젝트 루트의 `game-save-backup-manager.exe`.
- Windows 패키징은 `FyneApp.toml` 메타데이터 기반 `fyne package -os windows` 사용.
- `fyne package -os windows -release`는 배포 산출물이 필요할 때만 사용.
- `fyne package` 실행 후에는 항상 `git restore -- FyneApp.toml`로 `Build` 자동 증가 원복.
- `FyneApp.toml`의 `Build` 기준값은 `1` 유지.
- 기능/설정/계약 변경 시 `README.md`와 `AGENTS.md`를 함께 갱신.
- Git 커밋은 작업 완료 후 에이전트가 자율 수행하되, 의미 있는 기능/수정 단위로 분리.

## 데이터 운영 정책 (중요)
- `savelocation.db`는 **실제 데이터**이며, 실제 데이터 포함 배포가 의도다.
- 따라서 `savelocation.db`를 임의로 삭제/초기화/샘플 DB로 치환하지 않는다.
- DB 스키마: `savelocation(name,type,filename,path)`, PK/UNIQUE 없음, 수정/삭제는 `rowid` 기준.

## 계약상 고정 동작 (깨지기 쉬움)
- `reg` 경로 해석 규칙은 고정:
  - 경로가 `\\`로 끝나면 키 전체를 대상(백업/검증).
  - 경로가 `\\`로 끝나지 않으면 마지막 세그먼트를 값 이름으로 간주(값 단일 백업/검증).
- 위 `reg` 규칙은 스캔/백업/입력검증에 동일 적용.
- `reg` 타입의 `경로 열기`는 파일 탐색기가 아니라 레지스트리 편집기 실행.
- 레지스트리 편집기 실행 실패 시 `RunAs` 승격 재시도.
- `zip` 백업은 ZIP 내부 디렉터리 계층 보존 + 플레이스홀더 기반 논리 경로 유지.

## 플레이스홀더 규칙
- 지원 토큰:
  - `{{steam-path}}`
  - `{{steam-userid}}`
  - `{{steam-accountid}}`
  - `{{microsoftstore-userid}}`
  - `{{rockstargameslauncher-userid}}`
  - `{{ubisoftconnect-path}}`
  - `{{ubisoftconnect-userid}}`
- 스캔/실제 파일 접근 시에는 설정값으로 치환.
- ZIP 내부 논리 경로에는 토큰 원문 유지.
- 치환에 필요한 설정값이 비어 있으면 해당 항목은 미감지 처리될 수 있음.

## 설정 파일 규칙
- 형식: INI
- 파일명: `<exe basename>.ini`
- 위치: 실행 파일과 같은 디렉터리
- 키:
  - `steam_path`
  - `steam_userid`
  - `steam_accountid`
  - `ubisoft_connect_path`
  - `ubisoft_connect_userid`
  - `rockstargameslauncher_userid`
  - `microsoftstore_userid`
- ini가 없으면 시작 시 자동 생성.
- 기본 경로:
  - Steam: `%PROGRAMFILES(X86)%\Steam`
  - Ubisoft Connect: `%PROGRAMFILES(X86)%\Ubisoft\Ubisoft Game Launcher`
- UI 표시/저장 시 환경변수 경로는 실제 경로로 치환해 사용.

## UI 계약 (의도 보존용)
- 메인 상단: 좌 `스캔`, 우 `설정`.
- `설정`은 별도 창이 아닌 다이얼로그.
- 설정 다이얼로그 탭: `Steam`, `Ubisoft`, `Rockstar Games`, `Microsoft`.
- 탭별 필드:
  - Steam: 설치 경로 + USER ID + ACCOUNT ID
  - Ubisoft: 설치 경로 + USER ID
  - Rockstar Games: USER ID만
  - Microsoft: USER ID만
- 각 경로 입력행 우측 `폴더 선택` 버튼.
- 폼 바깥 `DB 관리` 버튼.
- 하단 버튼: `취소` / `저장` (1:1 폭, 저장 강조).
- 설정 화면은 `dialog.NewForm`이 아니라 `NewCustomWithoutButtons` 사용.
  - 이유: "라벨+입력 2줄 고정"과 "DB 관리 버튼 폼 바깥 배치"를 동시에 만족해야 하기 때문.
- 메인 목록은 전체 DB가 아니라 감지된 세이브만 표시.
- 메인 목록 행 구성: 게임명 + `백업` + `경로 열기`.
- 메인 창 종료 시 열려 있는 DB 관리 창도 함께 닫음.

## 빌드/검증 절차
1. `go build ./...`
2. 패키징 검증(로컬 확인): `fyne package -os windows`
   - 직후 `git restore -- FyneApp.toml`
3. 배포 빌드(필요 시만): `fyne package -os windows -release`
   - 직후 `git restore -- FyneApp.toml`

## 환경 특이사항
- 이 환경은 `go build`/`go list` 계열에서 캐시 경로 권한 이슈가 간헐적으로 발생.
- 권한 문제로 실패하면 동일 명령을 권한 상승으로 재실행 시 통과되는 경우가 많음.
- 현재 프로젝트는 Git 저장소.
