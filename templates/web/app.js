const form = document.getElementById('acta-form');
const elementosContainer = document.getElementById('elementos-container');
const resultado = document.getElementById('resultado');

const tipoSelect = document.getElementById('tipo');
const campoQuienEntrega = document.getElementById('quienEntrega');
const campoQuienRecibe = document.getElementById('quienRecibe');
const labelQuienEntrega = campoQuienEntrega.closest('label');
const labelQuienRecibe = campoQuienRecibe.closest('label');

// direccionForzada devuelve la única dirección válida de elemento para un
// tipo de acta que no sea "ambos", o null si ambas direcciones son válidas.
function direccionForzada(tipo) {
  if (tipo === 'entrega') return 'retira';
  if (tipo === 'recibimiento') return 'entrega';
  return null;
}

// aplicarDireccionEnFila oculta/muestra el selector de dirección de una fila
// y fija su valor cuando el tipo de acta solo admite una dirección.
function aplicarDireccionEnFila(fila, tipo) {
  const selectDireccion = fila.querySelector('.elemento-direccion');
  const forzada = direccionForzada(tipo);
  if (forzada) {
    selectDireccion.value = forzada;
    selectDireccion.hidden = true;
  } else {
    selectDireccion.hidden = false;
  }
}

// actualizarCamposPorTipo muestra/oculta y (des)marca required en los campos
// de persona según el tipo elegido, y resincroniza la dirección de todas las
// filas de elementos ya creadas.
function actualizarCamposPorTipo() {
  const tipo = tipoSelect.value;
  const mostrarEntrega = tipo === 'entrega' || tipo === 'ambos';
  const mostrarRecibe = tipo === 'recibimiento' || tipo === 'ambos';

  // QuienEntrega conserva su valor (por defecto "Oficina de Sistemas") al
  // ocultarse: no tiene sentido pedirlo de nuevo cada vez que se vuelve a
  // mostrar, a diferencia de QuienRecibe que sí se vacía.
  labelQuienEntrega.hidden = !mostrarEntrega;
  campoQuienEntrega.required = mostrarEntrega;

  labelQuienRecibe.hidden = !mostrarRecibe;
  campoQuienRecibe.required = mostrarRecibe;
  if (!mostrarRecibe) campoQuienRecibe.value = '';

  elementosContainer.querySelectorAll('.elemento-row').forEach((fila) => {
    aplicarDireccionEnFila(fila, tipo);
  });
}

tipoSelect.addEventListener('change', actualizarCamposPorTipo);

function crearFilaElemento() {
  const fila = document.createElement('div');
  fila.className = 'elemento-row';
  fila.innerHTML = `
    <input type="text" class="elemento-descripcion" placeholder="Descripción" required>
    <input type="text" class="elemento-nroInventario" placeholder="Nro. Inventario">
    <input type="number" class="elemento-cantidad" placeholder="Cantidad" min="1" value="1" required>
    <select class="elemento-direccion">
      <option value="retira">Retira</option>
      <option value="entrega">Entrega</option>
    </select>
    <input type="text" class="elemento-observacion" placeholder="Observación">
    <button type="button" class="quitar-elemento">Quitar</button>
  `;
  fila.querySelector('.quitar-elemento').addEventListener('click', () => fila.remove());
  elementosContainer.appendChild(fila);
  aplicarDireccionEnFila(fila, tipoSelect.value);
}

document.getElementById('agregar-elemento').addEventListener('click', crearFilaElemento);

crearFilaElemento();
actualizarCamposPorTipo();

function leerElementos() {
  return Array.from(elementosContainer.querySelectorAll('.elemento-row')).map((fila) => ({
    descripcion: fila.querySelector('.elemento-descripcion').value,
    nroInventario: fila.querySelector('.elemento-nroInventario').value,
    cantidad: Number(fila.querySelector('.elemento-cantidad').value),
    direccion: fila.querySelector('.elemento-direccion').value,
    observacion: fila.querySelector('.elemento-observacion').value,
  }));
}

// mostrarToast crea una notificación flotante en la esquina superior
// derecha, con fade-in/translateY de entrada, y la retira sola a los 3s.
function mostrarToast(mensaje, tipo) {
  const toast = document.createElement('div');
  toast.className = `toast toast-${tipo}`;
  toast.textContent = mensaje;
  document.body.appendChild(toast);

  requestAnimationFrame(() => {
    toast.classList.add('toast-visible');
  });

  setTimeout(() => {
    toast.classList.remove('toast-visible');
    toast.addEventListener('transitionend', () => toast.remove(), { once: true });
  }, 3000);
}

form.addEventListener('submit', async (event) => {
  event.preventDefault();
  resultado.textContent = '';

  const fecha = document.getElementById('fecha').value;
  const acta = {
    fecha: fecha ? `${fecha}T00:00:00Z` : '',
    quienEntrega: document.getElementById('quienEntrega').value,
    quienRecibe: document.getElementById('quienRecibe').value,
    tipo: document.getElementById('tipo').value,
    areaDepartamento: document.getElementById('areaDepartamento').value,
    elementos: leerElementos(),
  };

  try {
    const respuesta = await fetch('/api/actas/generar', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(acta),
    });
    const datos = await respuesta.json();

    if (!respuesta.ok) {
      throw new Error(datos.error || 'Error desconocido');
    }

    resultado.textContent = '';

    const enlace = document.createElement('a');
    enlace.href = datos.descargaUrl;
    enlace.textContent = 'Descargar acta';
    enlace.setAttribute('download', '');
    resultado.appendChild(enlace);

    mostrarToast(datos.mensaje, 'exito');
  } catch (error) {
    resultado.textContent = '';
    mostrarToast(`Error: ${error.message}`, 'error');
  }
});
