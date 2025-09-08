# GitOps Validator Release Script for PowerShell
# Usage: .\scripts\release.ps1 [version] [message]
# Example: .\scripts\release.ps1 v1.0.0 "Initial release"

param(
    [Parameter(Position=0, Mandatory=$true)]
    [string]$Version,
    
    [Parameter(Position=1, Mandatory=$true)]
    [string]$Message,
    
    [switch]$Help
)

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Function to show usage
function Show-Usage {
    Write-Host "Usage: .\scripts\release.ps1 [version] [message]"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\scripts\release.ps1 v1.0.0 'Initial release'"
    Write-Host "  .\scripts\release.ps1 v1.1.0 'Bug fixes and improvements'"
    Write-Host "  .\scripts\release.ps1 v2.0.0 'Major release with new features'"
    Write-Host ""
    Write-Host "Version format: vX.Y.Z (e.g., v1.0.0)"
    Write-Host "Message: Description of the release"
}

# Function to check if we're in a git repository
function Test-GitRepository {
    try {
        git rev-parse --git-dir | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Function to check if tag already exists
function Test-TagExists {
    param([string]$Tag)
    try {
        git rev-parse $Tag | Out-Null
        return $true
    }
    catch {
        return $false
    }
}

# Function to validate version format
function Test-VersionFormat {
    param([string]$Version)
    return $Version -match '^v\d+\.\d+\.\d+$'
}

# Function to check for uncommitted changes
function Test-UncommittedChanges {
    try {
        git diff-index --quiet HEAD --
        return $false
    }
    catch {
        return $true
    }
}

# Main execution
if ($Help) {
    Show-Usage
    exit 0
}

Write-Status "Starting release process for $Version..."

# Validate inputs
if (-not (Test-VersionFormat $Version)) {
    Write-Error "Invalid version format! Use format: v1.0.0"
    exit 1
}

# Check prerequisites
if (-not (Test-GitRepository)) {
    Write-Error "Not in a git repository!"
    exit 1
}

if (Test-TagExists $Version) {
    Write-Error "Tag $Version already exists!"
    exit 1
}

if (Test-UncommittedChanges) {
    Write-Warning "You have uncommitted changes!"
    $response = Read-Host "Do you want to continue anyway? (y/N)"
    if ($response -notmatch '^[Yy]$') {
        Write-Error "Release cancelled."
        exit 1
    }
}

# Update version
Write-Status "Updating version to $Version..."
Write-Status "Version $Version ready for release"

# Create release
Write-Status "Creating release $Version..."

# Add all changes
git add .

# Commit changes
try {
    git commit -m "Prepare for release $Version"
}
catch {
    Write-Warning "No changes to commit"
}

# Create tag
git tag -a $Version -m $Message

# Push changes and tag
git push origin main
git push origin $Version

Write-Success "Release $Version created and pushed!"
Write-Status "GitHub Actions will now build and upload the release assets."
Write-Status "Check the Actions tab in your GitHub repository to monitor progress."

# Get repository URL for release link
try {
    $repoUrl = git config --get remote.origin.url
    $repoUrl = $repoUrl -replace '.*github.com[:/]([^.]*).*', '$1'
    $repoUrl = $repoUrl -replace '\.git$', ''
    $repoUrl = $repoUrl -replace 'git@github.com:', ''
    $repoUrl = $repoUrl -replace 'https://github.com/', ''
    
    Write-Success "Release process completed!"
    Write-Status "Release URL: https://github.com/$repoUrl/releases/tag/$Version"
}
catch {
    Write-Success "Release process completed!"
    Write-Status "Check your GitHub repository for the new release."
}
