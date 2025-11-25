Add-Type @"
using System;
using System.Runtime.InteropServices;

public class Win32 {
    [DllImport("user32.dll")]
    public static extern IntPtr GetForegroundWindow();

    [DllImport("user32.dll")]
    public static extern bool GetWindowRect(IntPtr hWnd, out RECT rect);

    [StructLayout(LayoutKind.Sequential)]
    public struct RECT {
        public int Left;
        public int Top;
        public int Right;
        public int Bottom;
    }
}
"@

$now = Get-Date -Format "yyyyMMdd_HHmmss"
$outDir = Join-Path $PSScriptRoot 'screenshots'
if (-not (Test-Path $outDir)) { New-Item -ItemType Directory -Path $outDir | Out-Null }

$h = [Win32]::GetForegroundWindow()
if ($h -eq [IntPtr]::Zero) {
    Write-Error "无法获取前台窗口句柄。请把微信开发者工具窗口置于最前并重试。"
    exit 1
}

[Win32+RECT]$r = New-Object Win32+RECT
if (-not [Win32]::GetWindowRect($h,[ref]$r)) {
    Write-Error "获取窗口位置失败"
    exit 1
}

$width = $r.Right - $r.Left
$height = $r.Bottom - $r.Top

Add-Type -AssemblyName System.Drawing
$bmp = New-Object System.Drawing.Bitmap $width, $height
$g = [System.Drawing.Graphics]::FromImage($bmp)
$g.CopyFromScreen($r.Left, $r.Top, 0, 0, $bmp.Size)

$outFile = Join-Path $outDir ("wechat_" + $now + ".png")
$bmp.Save($outFile, [System.Drawing.Imaging.ImageFormat]::Png)
$g.Dispose()
$bmp.Dispose()

Write-Output "Saved screenshot to: $outFile"