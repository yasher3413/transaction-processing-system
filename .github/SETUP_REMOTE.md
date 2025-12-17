# Fix GitHub Remote URL

## Option 1: Use HTTPS (Easiest - No SSH Setup)

Replace `YOUR_USERNAME` and `REPO_NAME` with your actual values:

```bash
git remote set-url origin https://github.com/YOUR_USERNAME/REPO_NAME.git
git push -u origin main
```

You'll be prompted for your GitHub username and password (or personal access token).

## Option 2: Set Up SSH Keys

If you prefer SSH:

1. **Generate SSH key:**
   ```bash
   ssh-keygen -t ed25519 -C "your_email@example.com"
   # Press Enter to accept default location
   # Press Enter twice for no passphrase (or set one)
   ```

2. **Add to SSH agent:**
   ```bash
   eval "$(ssh-agent -s)"
   ssh-add ~/.ssh/id_ed25519
   ```

3. **Copy public key:**
   ```bash
   cat ~/.ssh/id_ed25519.pub
   ```

4. **Add to GitHub:**
   - Go to https://github.com/settings/keys
   - Click "New SSH key"
   - Paste the key
   - Click "Add SSH key"

5. **Update remote and push:**
   ```bash
   git remote set-url origin git@github.com:YOUR_USERNAME/REPO_NAME.git
   git push -u origin main
   ```

## Quick Fix (If username is yasher3413)

```bash
git remote set-url origin https://github.com/yasher3413/transaction-processing-system.git
git push -u origin main
```

