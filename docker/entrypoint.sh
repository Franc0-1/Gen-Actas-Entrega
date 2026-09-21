#!/bin/sh
set -eu

# Mismo perfil que usa internal/generator/pdf.go (os.TempDir() + el mismo
# nombre de carpeta) para las conversiones puntuales: al compartir perfil,
# soffice detecta esta instancia ya activa y le delega la conversión en vez
# de arrancar una nueva, evitando el costo de boot en cada acta generada.
PROFILE_DIR="${TMPDIR:-/tmp}/actas-libreoffice-profile"
PROFILE_URI="file://$PROFILE_DIR"

supervisar_soffice() {
  while true; do
    soffice --headless --invisible --nologo --nodefault --norestore \
      --nofirststartwizard \
      "-env:UserInstallation=$PROFILE_URI" &
    soffice_pid=$!
    wait "$soffice_pid"
    echo "entrypoint: soffice se detuvo, reiniciando en 2s..." >&2
    sleep 2
  done
}

supervisar_soffice &

exec ./server
