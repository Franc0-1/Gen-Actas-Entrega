const form = document.getElementById('acta-form');
const elementosContainer = document.getElementById('elementos-container');

const tipoSelect = document.getElementById('tipo');
const campoQuienEntrega = document.getElementById('quienEntrega');
const campoQuienRecibe = document.getElementById('quienRecibe');
const labelQuienEntrega = campoQuienEntrega.closest('label');
const labelQuienRecibe = campoQuienRecibe.closest('label');

const campoFecha = document.getElementById('fecha');
const headerDireccion = document.querySelector('.elementos-header .col-direccion');

// Por defecto la fecha es la de hoy (fecha local del navegador); el
// usuario sigue pudiendo cambiarla, no queda bloqueada.
function fechaHoyISO() {
  const hoy = new Date();
  const yyyy = hoy.getFullYear();
  const mm = String(hoy.getMonth() + 1).padStart(2, '0');
  const dd = String(hoy.getDate()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

campoFecha.value = fechaHoyISO();

// soloDigitos limpia cualquier carácter que no sea 0-9, para usar en
// campos que el usuario podría pegar o escribir con caracteres inválidos.
function soloDigitos(valor) {
  return valor.replace(/[^0-9]/g, '');
}

// construirFechaISO toma el valor del date picker nativo ("AAAA-MM-DD") y
// arma el RFC3339 que espera el backend, o null si está vacío. El input
// type="date" ya solo permite fechas calendario válidas.
function construirFechaISO() {
  if (!campoFecha.value) return null;
  return `${campoFecha.value}T00:00:00Z`;
}

// sanearCantidad fuerza que .elemento-cantidad quede en un entero >= 1,
// aunque el usuario haya escrito un decimal, negativo o texto inválido.
function sanearCantidad(input) {
  if (input.value === '') return;
  let n = Math.floor(Number(input.value));
  if (!Number.isFinite(n) || n < 1) n = 1;
  input.value = String(n);
}

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

  // Misma condición que aplicarDireccionEnFila: si el tipo fuerza una
  // única dirección, la columna de la cabecera se oculta igual que el
  // selector de cada fila, para que ambas sigan alineadas.
  headerDireccion.hidden = Boolean(direccionForzada(tipo));

  elementosContainer.querySelectorAll('.elemento-row').forEach((fila) => {
    aplicarDireccionEnFila(fila, tipo);
  });
}

tipoSelect.addEventListener('change', actualizarCamposPorTipo);

// crearFilaElemento arma una fila nueva. Si se le pasan datosIniciales
// (usado al duplicar), prellena descripción/nroInventario/observación/
// dirección con esos valores. Devuelve la fila creada.
function crearFilaElemento(datosIniciales) {
  const fila = document.createElement('div');
  fila.className = 'elemento-row';
  fila.innerHTML = `
    <input type="text" class="elemento-descripcion" placeholder="Descripción" required>
    <input type="text" class="elemento-nroInventario" placeholder="Nro. Inventario" inputmode="numeric" pattern="[0-9]*">
    <input type="number" class="elemento-cantidad" placeholder="Cantidad" min="1" step="1" inputmode="numeric" value="1" required>
    <select class="elemento-direccion">
      <option value="retira">Retira</option>
      <option value="entrega">Entrega</option>
    </select>
    <input type="text" class="elemento-observacion" placeholder="Observación">
    <div class="elemento-acciones">
      <button type="button" class="duplicar-elemento">Duplicar</button>
      <button type="button" class="quitar-elemento">Quitar</button>
    </div>
  `;

  if (datosIniciales) {
    fila.querySelector('.elemento-descripcion').value = datosIniciales.descripcion || '';
    fila.querySelector('.elemento-nroInventario').value = datosIniciales.nroInventario || '';
    fila.querySelector('.elemento-observacion').value = datosIniciales.observacion || '';
    if (datosIniciales.direccion) {
      fila.querySelector('.elemento-direccion').value = datosIniciales.direccion;
    }
  }

  elementosContainer.appendChild(fila);
  aplicarDireccionEnFila(fila, tipoSelect.value);

  const campoNroInventario = fila.querySelector('.elemento-nroInventario');
  campoNroInventario.addEventListener('input', () => {
    campoNroInventario.value = soloDigitos(campoNroInventario.value);
  });

  const campoCantidad = fila.querySelector('.elemento-cantidad');
  campoCantidad.addEventListener('input', () => sanearCantidad(campoCantidad));

  fila.querySelector('.duplicar-elemento').addEventListener('click', () => {
    agregarFilaYEnfocar({
      descripcion: fila.querySelector('.elemento-descripcion').value,
      nroInventario: fila.querySelector('.elemento-nroInventario').value,
      observacion: fila.querySelector('.elemento-observacion').value,
      direccion: fila.querySelector('.elemento-direccion').value,
    });
  });

  fila.querySelector('.quitar-elemento').addEventListener('click', () => eliminarFila(fila));

  return fila;
}

// agregarFilaYEnfocar crea una fila (con datosIniciales opcionales, para
// duplicar) y pone el foco en su primer campo, para poder seguir cargando
// sin usar el mouse.
function agregarFilaYEnfocar(datosIniciales) {
  const fila = crearFilaElemento(datosIniciales);
  fila.querySelector('.elemento-descripcion').focus();
  return fila;
}

// eliminarFila saca una fila del formulario. Si hay más de una, pide
// confirmación; si es la única, la borra directo. Después mueve el foco a
// una fila vecina o, si no queda ninguna, al botón de agregar.
function eliminarFila(fila) {
  const filas = Array.from(elementosContainer.querySelectorAll('.elemento-row'));
  if (filas.length > 1 && !confirm('¿Eliminar este elemento?')) return;

  const indice = filas.indexOf(fila);
  const filaVecina = filas[indice - 1] || filas[indice + 1];

  fila.remove();

  if (filaVecina) {
    filaVecina.querySelector('.elemento-descripcion').focus();
  } else {
    document.getElementById('agregar-elemento').focus();
  }
}

document.getElementById('agregar-elemento').addEventListener('click', () => agregarFilaYEnfocar());

// Enter en el último campo (Observación) de la última fila agrega una fila
// nueva en vez de enviar el formulario. En cualquier otro campo, Enter no
// se intercepta y se comporta como ya se comportaba.
elementosContainer.addEventListener('keydown', (event) => {
  if (event.key !== 'Enter') return;
  if (!event.target.classList.contains('elemento-observacion')) return;

  const fila = event.target.closest('.elemento-row');
  const filas = elementosContainer.querySelectorAll('.elemento-row');
  if (filas[filas.length - 1] !== fila) return;

  event.preventDefault();
  agregarFilaYEnfocar();
});

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
// Si se pasa enlaceDescarga, agrega un link breve dentro del toast: sirve
// de respaldo si el navegador bloqueó la descarga automática (no hay forma
// confiable de detectar el bloqueo, así que siempre se ofrece).
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

  requestAnimationFrame(() => {
    toast.classList.add('toast-visible');
  });

  setTimeout(() => {
    toast.classList.remove('toast-visible');
    toast.addEventListener('transitionend', () => toast.remove(), { once: true });
  }, 3000);
}

const botonGenerar = form.querySelector('button[type="submit"]');
const textoOriginalBoton = botonGenerar.textContent;
let generandoEnCurso = false;

form.addEventListener('submit', async (event) => {
  event.preventDefault();

  // Guarda explícita contra doble envío, además del disabled del botón:
  // cubre también un submit disparado por Enter mientras ya hay uno en
  // curso.
  if (generandoEnCurso) return;

  
  const fecha = construirFechaISO();
  if (!fecha) {
    mostrarToast('La fecha ingresada no es válida', 'error');
    return;
  }

  const acta = {
    fecha,
    quienEntrega: document.getElementById('quienEntrega').value,
    quienRecibe: document.getElementById('quienRecibe').value,
    tipo: document.getElementById('tipo').value,
    areaDepartamento: document.getElementById('areaDepartamento').value,
    elementos: leerElementos(),
  };

  generandoEnCurso = true;
  botonGenerar.disabled = true;
  botonGenerar.innerHTML = '<span class="spinner"></span> Generando PDF...';

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

    // Descarga automática: el <a> ni siquiera hace falta insertarlo en el
    // DOM para que el click dispare la descarga.
    const enlace = document.createElement('a');
    enlace.href = datos.descargaUrl;
    enlace.setAttribute('download', '');
    enlace.click();

    mostrarToast(datos.mensaje, 'exito', datos.descargaUrl);
  } catch (error) {
    mostrarToast(`Error: ${error.message}`, 'error');
  } finally {
    generandoEnCurso = false;
    botonGenerar.disabled = false;
    botonGenerar.textContent = textoOriginalBoton;
  }
});
