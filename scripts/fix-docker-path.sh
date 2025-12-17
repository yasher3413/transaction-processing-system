#!/bin/bash
# Fix Docker PATH issue on macOS

echo "Fixing Docker PATH..."

# Check if Docker exists
if [ -f "/Applications/Docker.app/Contents/Resources/bin/docker" ]; then
    echo "Docker found at /Applications/Docker.app/Contents/Resources/bin/docker"
    
    # Add to PATH in .zshrc (since you're using zsh)
    if ! grep -q "Docker.app/Contents/Resources/bin" ~/.zshrc 2>/dev/null; then
        echo "" >> ~/.zshrc
        echo "# Docker Desktop PATH" >> ~/.zshrc
        echo 'export PATH="/Applications/Docker.app/Contents/Resources/bin:$PATH"' >> ~/.zshrc
        echo "✅ Added Docker to PATH in ~/.zshrc"
        echo "📝 Run: source ~/.zshrc (or restart terminal)"
    else
        echo "✅ Docker PATH already in ~/.zshrc"
    fi
    
    # Also fix the broken symlink (requires sudo)
    if [ -L "/usr/local/bin/docker" ] && [ ! -e "/usr/local/bin/docker" ]; then
        echo ""
        echo "⚠️  Found broken symlink at /usr/local/bin/docker"
        echo "To fix it, run:"
        echo "  sudo rm /usr/local/bin/docker"
        echo "  sudo ln -s /Applications/Docker.app/Contents/Resources/bin/docker /usr/local/bin/docker"
    fi
else
    echo "❌ Docker not found at expected location"
    echo "Make sure Docker Desktop is installed"
fi

