# Git Helpers - Feature Branch Management

Quick tools to create and manage feature branches with your GitHub username as a prefix.

## Setup (One-time)

Run this once to configure git aliases and hooks:

```bash
bash .githooks/setup-git-helpers.sh
```

This will:
- Set `.githooks` as the git hooks directory
- Create git aliases for quick branch creation
- Enable the custom helper scripts

## Usage

### Option 1: GitHub Actions (Remote)

Trigger the workflow from GitHub:

1. Go to **Actions** → **Create Feature Branch**
2. Click **Run workflow**
3. Enter:
   - **Branch name**: `my-feature` (will become `username/my-feature`)
   - **Base branch**: `main` or `develop`
4. Click **Run workflow**
5. Checkout locally:
   ```bash
   git fetch origin
   git checkout username/my-feature
   ```

### Option 2: Local Helper Script (Recommended)

After setup, use the git aliases:

```bash
# Create from current branch
git feature my-feature

# Create from main branch
git feature-main my-feature

# Create from develop branch
git feature-develop my-feature
```

This will:
1. ✅ Create branch named `<your-username>/my-feature`
2. ✅ Automatically check out that branch
3. ✅ Set upstream tracking
4. ✅ Show next steps

### Option 3: Manual Script

Run the helper script directly:

```bash
bash .githooks/create-feature-branch.sh my-feature main
```

## Examples

```bash
# Create a feature branch for authentication
git feature-main add-bcrypt-hashing

# Branch created: paul/add-bcrypt-hashing
# Checked out and ready to work!
```

## Branch Naming Convention

All branches follow the pattern: `<github-username>/<feature-name>`

Examples:
- `paul/add-authentication`
- `alice/fix-database-connection`
- `bob/update-logging-system`

## Git Aliases Created

| Alias | Command |
|-------|---------|
| `git feature` | Create branch from current branch |
| `git feature-main` | Create branch from `main` |
| `git feature-develop` | Create branch from `develop` |

## Next Steps After Creating a Branch

```bash
# Make your changes
git add .
git commit -m "Your meaningful commit message"

# Push to remote (sets upstream tracking)
git push -u origin

# Or if already pushed:
git push

# Create a Pull Request on GitHub
```

## Troubleshooting

**Q: "Could not determine GitHub username"**
- Ensure your git config has a `user.name` set:
  ```bash
  git config --global user.name "your-username"
  ```

**Q: Branch creation failed**
- Ensure you have permission to push to the repository
- Check that `origin` remote is configured:
  ```bash
  git remote -v
  ```

**Q: Want to change base branch after creation?**
- Recreate with the correct base:
  ```bash
  git checkout main  # or develop
  git pull origin main
  git checkout -b username/new-feature
  ```

## Customization

Edit the aliases in `setup-git-helpers.sh` or configure manually:

```bash
# Create custom alias
git config alias.my-alias '!bash .githooks/create-feature-branch.sh $1 develop'
```

View all configured aliases:
```bash
git config --get-regexp alias
```
