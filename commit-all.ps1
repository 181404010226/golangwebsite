# 获取提交信息
$commitMessage = Read-Host -Prompt 'Enter commit message'

# 如果没有输入提交信息，则使用默认信息
if ([string]::IsNullOrWhiteSpace($commitMessage)) {
    $commitMessage = "Update both main and submodule repositories"
}

Write-Host "`nProcessing submodule (meeting-app)..." -ForegroundColor Cyan

# 进入子模块目录
Push-Location meeting-app

# 检查子模块是否有修改
$hasSubmoduleChanges = git status --porcelain
if ($hasSubmoduleChanges) {
    # 处理子模块
    git add .
    git commit -m "$commitMessage"
    git push
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Submodule changes pushed successfully`n" -ForegroundColor Green
    } else {
        Write-Host "Error pushing submodule changes`n" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "No changes in submodule to commit`n" -ForegroundColor Yellow
}

# 返回主目录
Pop-Location

Write-Host "Processing main repository..." -ForegroundColor Cyan

# 检查主仓库是否有修改
$hasMainChanges = git status --porcelain
if ($hasMainChanges) {
    # 处理主仓库
    git add .
    git commit -m "$commitMessage"
    git push
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Main repository changes pushed successfully" -ForegroundColor Green
    } else {
        Write-Host "Error pushing main repository changes" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "No changes in main repository to commit" -ForegroundColor Yellow
}

Write-Host "`nAll changes have been committed and pushed successfully!" -ForegroundColor Green