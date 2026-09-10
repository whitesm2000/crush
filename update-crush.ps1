# update-crush.ps1 - Pull upstream Crush, rebase local features, rebuild.
# Run from anywhere:  powershell -File "$HOME\crush-src\update-crush.ps1"
param(
    [string]$DeployTo = ""  # optional: full path of crush.exe to replace with the fresh build
)

$ErrorActionPreference = "Stop"
Set-Location "$env:USERPROFILE\crush-src"
$Go = "C:\Program Files\Go\bin\go.exe"

Write-Host "== Fetching upstream ==" -ForegroundColor Cyan
git checkout main
git pull upstream main

Write-Host "`n== Local commits upstream does NOT have (your keep-list) ==" -ForegroundColor Cyan
$keep = git log --cherry-pick --right-only --oneline "main...mine"
if ($keep) { $keep | Write-Host } else { Write-Host "(none - mine matches upstream, nothing custom to carry)" }

Write-Host "`n== Rebasing feature branches onto new main ==" -ForegroundColor Cyan
$features = git branch --list "feat/*" "local-*" --format="%(refname:short)"
foreach ($f in $features) {
    git checkout $f
    git rebase main
    if ($LASTEXITCODE -ne 0) {
        Write-Host "CONFLICT rebasing $f - resolve manually, then 'git rebase --continue' and re-run this script." -ForegroundColor Red
        exit 1
    }
}

Write-Host "`n== Rebuilding integration branch 'mine' ==" -ForegroundColor Cyan
git checkout main
git branch -f mine main
git checkout mine
foreach ($f in $features) { git merge --no-ff $f -m "Merge $f into mine" }

Write-Host "`n== Building ==" -ForegroundColor Cyan
& $Go build -o "$env:USERPROFILE\crush-src\crush-patched.exe" .
if ($LASTEXITCODE -ne 0) { Write-Host "BUILD FAILED" -ForegroundColor Red; exit 1 }

if ($DeployTo -ne "") {
    Copy-Item "$env:USERPROFILE\crush-src\crush-patched.exe" $DeployTo -Force
    Write-Host "Deployed to $DeployTo" -ForegroundColor Green
}

Write-Host "`nDone. Reminder: update LOCAL-FEATURES.md if features were added/removed." -ForegroundColor Green
git log --oneline -5
