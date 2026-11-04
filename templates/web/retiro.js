// Tema: misma clave y lógica que app.js, para que la preferencia se comparta.
const CLAVE_TEMA = 'actas-tema';
const botonTema = document.getElementById('theme-toggle');

function aplicarTema(esOscuro) {
  document.body.classList.toggle('dark-theme', esOscuro);
  botonTema.textContent = esOscuro ? '☀︎' : '⏾';
  botonTema.setAttribute('aria-label', esOscuro ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro');
}

let temaGuardado = null;
try { temaGuardado = localStorage.getItem(CLAVE_TEMA); } catch (error) { /* sin localStorage */ }
aplicarTema(temaGuardado ? temaGuardado === 'oscuro' : window.matchMedia('(prefers-color-scheme: dark)').matches);

botonTema.addEventListener('click', () => {
  const esOscuro = !document.body.classList.contains('dark-theme');
  aplicarTema(esOscuro);
  try { localStorage.setItem(CLAVE_TEMA, esOscuro ? 'oscuro' : 'claro'); } catch (error) { /* sin localStorage */ }
});

const form = document.getElementById('retiro-form');
const campoFecha = document.getElementById('fecha');
const elementosContainer = document.getElementById('elementos-container');

const hoy = new Date();
campoFecha.value = `${hoy.getFullYear()}-${String(hoy.getMonth() + 1).padStart(2, '0')}-${String(hoy.getDate()).padStart(2, '0')}`;

function agregarElemento() {
  const fila = document.createElement('div');
  fila.className = 'elemento-row elemento-row--simple';
  fila.innerHTML = `
    <input type="text" class="elemento-descripcion" placeholder="Ej. Impresora marca Pantum modelo M6550NW nro. de serie CJ3X001670" aria-label="Descripción del elemento" required>
    <div class="elemento-acciones">
      <button type="button" class="quitar-elemento">Quitar</button>
    </div>
  `;
  fila.querySelector('.quitar-elemento').addEventListener('click', () => {
    if (elementosContainer.children.length > 1) fila.remove();
  });
  elementosContainer.appendChild(fila);
  fila.querySelector('input').focus();
}

document.getElementById('agregar-elemento').addEventListener('click', agregarElemento);
agregarElemento();
campoFecha.focus();

// Reutiliza los estilos de toast de style.css.
function mostrarToast(mensaje, tipo, enlaceDescarga) {
  const toast = document.createElement('div');
  toast.className = `toast toast-${tipo}`;
  toast.textContent = mensaje;
  if (enlaceDescarga) {
    const link = document.createElement('a');
    link.href = enlaceDescarga;
    link.textContent = 'Descargar';
    link.setAttribute('download', '');
    link.className = 'toast-link';
    toast.appendChild(link);
  }
  document.body.appendChild(toast);
  requestAnimationFrame(() => toast.classList.add('toast-visible'));
  setTimeout(() => {
    toast.classList.remove('toast-visible');
    toast.addEventListener('transitionend', () => toast.remove(), { once: true });
  }, 3000);
}

const botonGenerar = form.querySelector('button[type="submit"]');
const textoOriginalBoton = botonGenerar.textContent;

form.addEventListener('submit', async (event) => {
  event.preventDefault();
  if (botonGenerar.disabled) return;

  const nota = {
    fecha: `${campoFecha.value}T00:00:00Z`,
    empresa: document.getElementById('empresa').value,
    elementos: Array.from(elementosContainer.querySelectorAll('.elemento-descripcion')).map((i) => i.value),
  };

  botonGenerar.disabled = true;
  botonGenerar.innerHTML = '<span class="spinner"></span> Generando PDF...';
  try {
    const respuesta = await fetch('/api/notas-retiro/generar', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(nota),
    });
    const datos = await respuesta.json();
    if (!respuesta.ok) throw new Error(datos.error || 'Error desconocido');

    const enlace = document.createElement('a');
    enlace.href = datos.descargaUrl;
    enlace.setAttribute('download', '');
    enlace.click();
    mostrarToast(datos.mensaje, 'exito', datos.descargaUrl);
  } catch (error) {
    mostrarToast(`Error: ${error.message}`, 'error');
  } finally {
    botonGenerar.disabled = false;
    botonGenerar.textContent = textoOriginalBoton;
  }
});
