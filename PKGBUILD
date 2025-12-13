# Maintainer: Nomadcxx <noovie@gmail.com>
pkgname=searxng-rama
_pkgname=searxng
pkgver=r9128.920b402
pkgrel=1
pkgdesc="SearXNG RAMA Edition - Pre-configured SearXNG with custom theme and systemd service"
arch=('any')
url="https://github.com/Nomadcxx/searxng-RAMA"
license=('AGPL3')
depends=('python' 'systemd')
makedepends=('openssl' 'git' 'python-build' 'python-wheel' 'python-setuptools'  'python-msgspec')
optdepends=(
    'redis: Caching support for improved performance'
    'valkey: Alternative caching support'
    'libmagic: File type detection for uploads'
    'p7zip: Archive support for file upload'
)
provides=('searxng')
conflicts=('searx' 'searx-git' 'searxng')
backup=('opt/searxng-rama/searx/settings.yml' 'etc/systemd/system/searxng-rama.service')
install=${pkgname}.install

_giturl="https://github.com/searxng/searxng"
_gitbranch="master"
source=(git+$_giturl#branch=$_gitbranch
        git+https://github.com/Nomadcxx/searxng-RAMA.git
        'searxng-rama-settings.yml')
b2sums=('SKIP' 'SKIP'
        'SKIP')

pkgver() {
  cd $_pkgname
  printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short=7 HEAD)"
}

build() {
  cd $_pkgname

  # Use RAMA custom settings file for build
  export SEARXNG_SETTINGS_PATH="${srcdir}/searxng-rama-settings.yml"
  
  # Generate random secret key in our custom settings
  sed -i -e "s/ultrasecretkey/`openssl rand -hex 32`/g" "${SEARXNG_SETTINGS_PATH}"
  
  # Generate version info
  python -m searx.version freeze
  sed -i "s|GIT_URL =.*|GIT_URL = \"${_giturl}\"|g" searx/version_frozen.py
  sed -i "s|GIT_BRANCH =.*|GIT_BRANCH = \"${_gitbranch}\"|g" searx/version_frozen.py
  
  # generate the wheel using the system python-build
  python -m build --no-isolation --wheel
}

package() {
  cd $_pkgname

  # create virtual environment
  export PIP_DISABLE_PIP_VERSION_CHECK=1
  export PYTHONDONTWRITEBYTECODE=1
  python -m venv "$pkgdir/opt/searxng-rama/venv/"
  source "$pkgdir/opt/searxng-rama/venv/bin/activate"
  
  # install searxng and dependencies using pip
  "$pkgdir/opt/searxng-rama/venv/bin/pip" install --upgrade pip installer
  "$pkgdir/opt/searxng-rama/venv/bin/pip" install -r "requirements.txt"
  "$pkgdir/opt/searxng-rama/venv/bin/python" -m installer dist/*.whl

  # remove references to pkgdir
  find "$pkgdir/opt/searxng-rama/venv/bin" -maxdepth 1 -type f -exec sed -i "s|${pkgdir}/|/|g" {} +
  find "$pkgdir/opt/searxng-rama/venv/pyvenv.cfg" -maxdepth 1 -type f -exec sed -i "s|${pkgdir}/|/|g" {} +
  find "$pkgdir/opt/searxng-rama/venv" -type f -name "*.py[co]" -delete
  find "$pkgdir/opt/searxng-rama/venv" -type d -name "__pycache__" -delete 

  local _site_packages="$(python -c 'import site, os; print(os.path.relpath(site.getsitepackages()[0]))')"

  # exit virtual environment
  deactivate

  # Install RAMA custom settings (overwrite the default ones)
  install -Dm644 "${srcdir}/searxng-rama-settings.yml" "${pkgdir}/opt/searxng-rama/searx/settings.yml"
  
  # Install limiter config
  install -Dm644 "searx/limiter.toml" "${pkgdir}/opt/searxng-rama/searx/limiter.toml"
  
  # Install version info
  install -Dm644 "searx/version_frozen.py" "${pkgdir}/opt/searxng-rama/${_site_packages}/searx/version_frozen.py"
  
  # Install licenses
  install -Dm644 "LICENSE" "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
  install -Dm644 "${srcdir}/searxng-RAMA/LICENSE" "${pkgdir}/usr/share/licenses/${pkgname}/RAMA_LICENSE" 2>/dev/null || true
  
  # Install RAMA customizations from our repo
  if [ -d "${srcdir}/searxng-RAMA" ]; then
    cd "${srcdir}/searxng-RAMA"
    
    # Install custom RAMA theme
    install -dm755 "${pkgdir}/opt/searxng-rama/searxng_extra/theme"
    cp -r theme/. "${pkgdir}/opt/searxng-rama/searxng_extra/theme/"
    
    # Install RAMA branding assets
    install -dm755 "${pkgdir}/opt/searxng-rama/searx/static/themes/rama"
    cp -r assets/. "${pkgdir}/opt/searxng-rama/searx/static/themes/rama/"
    cp -r brand/. "${pkgdir}/opt/searxng-rama/searx/static/themes/rama/brand/"
  fi
  
  # Create executable wrapper
  install -d "$pkgdir/usr/bin"
  cat > "$pkgdir/usr/bin/searxng-rama-run" << 'EOF'
#!/bin/bash
export SEARXNG_SETTINGS_PATH=/opt/searxng-rama/searx/settings.yml
exec /opt/searxng-rama/venv/bin/python -m searx.webapp "$@"
EOF
  chmod +x "$pkgdir/usr/bin/searxng-rama-run"
  
  # Install systemd service file
  install -dm755 "${pkgdir}/etc/systemd/system"
  cat > "${pkgdir}/etc/systemd/system/searxng-rama.service" << EOF
[Unit]
Description=RAMA SearXNG
After=network.target

[Service]
Type=simple
User=searxng
WorkingDirectory=/opt/searxng-rama
Environment="SEARXNG_SETTINGS_PATH=/opt/searxng-rama/searx/settings.yml"
ExecStart=/usr/bin/searxng-rama-run
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
  
  # Install README
  install -Dm644 "${srcdir}/searxng-RAMA/README.md" "${pkgdir}/usr/share/doc/${pkgname}/README.md" 2>/dev/null || true
}
