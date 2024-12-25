# 进入前端目录
Set-Location -Path "meeting-app"

# 安装依赖
Write-Host "Installing dependencies..."
cnpm install

# 构建前端项目
Write-Host "Building React app..."
npm run build

# 返回根目录
Set-Location -Path ".."

# 构建 Go 程序
Write-Host "Building Go backend..."
$timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$outputName = "server_${timestamp}.exe"
go build -o $outputName

Write-Host "Build complete! Output: $outputName"

Write-Host "Build complete!"
pause