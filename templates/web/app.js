const CLAVE_TEMA = 'actas-tema';
const botonTema = document.getElementById('theme-toggle');

// aplicarTema alterna la clase que redefine las variables CSS de color
// (ver body.dark-theme en style.css) y actualiza el label del botón; el
// ícono sol/luna lo alterna el CSS según esa misma clase.
function aplicarTema(esOscuro) {
  document.body.classList.toggle('dark-theme', esOscuro);
  botonTema.setAttribute('aria-label', esOscuro ? 'Cambiar a modo claro' : 'Cambiar a modo oscuro');
}

// Sin preferencia guardada en localStorage, se usa prefers-color-scheme del
// sistema operativo/navegador.
function temaGuardadoEsOscuro() {
  let guardado = null;
  try {
    guardado = localStorage.getItem(CLAVE_TEMA);
  } catch (error) {
    guardado = null;
  }
  if (guardado === 'oscuro') return true;
  if (guardado === 'claro') return false;
  return window.matchMedia('(prefers-color-scheme: dark)').matches;
}

aplicarTema(temaGuardadoEsOscuro());

botonTema.addEventListener('click', () => {
  const esOscuro = !document.body.classList.contains('dark-theme');
  aplicarTema(esOscuro);
  try {
    localStorage.setItem(CLAVE_TEMA, esOscuro ? 'oscuro' : 'claro');
  } catch (error) {
    // localStorage no disponible (ej. navegación privada): el tema no
    // persiste entre sesiones, pero el toggle sigue funcionando.
  }
});

const form = document.getElementById('doc-form');
const elementosContainer = document.getElementById('elementos-container');

// El documento (acta o nota de retiro) se elige con los radios
// name="documento" de #doc-picker; el tipo de acta sigue siendo el select #tipo.
const docPicker = document.getElementById('doc-picker');
const documentoActual = () => form.elements.documento.value;
const esRetiro = () => documentoActual() === 'retiro';
const tipoSelect = document.getElementById('tipo');
const tipoActual = () => tipoSelect.value;
const campoQuienEntrega = document.getElementById('quienEntrega');
const labelQuienEntrega = campoQuienEntrega.closest('label');
const textoAreaDepartamento = document.getElementById('areaDepartamentoTexto');

// El campo "Quién recibe" se reemplazó por el área: su texto indica qué rol
// cumple el área según el tipo de acta.
const TEXTO_AREA_POR_TIPO = {
  entrega: 'Área/Departamento que recibe',
  recibimiento: 'Área/Departamento que entrega',
  ambos: 'Área/Departamento que entrega y recibe',
};

const campoFecha = document.getElementById('fecha');
const headerDireccion = document.querySelector('.items-head .col-direccion');

const selectAreaDepartamento = document.getElementById('areaDepartamento');
const campoAreaDepartamentoOtra = document.getElementById('areaDepartamentoOtra');
const labelAreaDepartamentoOtra = campoAreaDepartamentoOtra.closest('label');

// Con "Otros" seleccionado se pide el nombre puntual del área en un input
// aparte, que se manda como areaDepartamento en vez del valor del select.
function actualizarAreaDepartamentoOtra() {
  const esOtros = selectAreaDepartamento.value === 'otros';
  labelAreaDepartamentoOtra.hidden = !esOtros;
  campoAreaDepartamentoOtra.required = esOtros;
}

selectAreaDepartamento.addEventListener('change', actualizarAreaDepartamentoOtra);
actualizarAreaDepartamentoOtra();

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
    selectDireccion.closest('.campo').hidden = true;
  } else {
    selectDireccion.closest('.campo').hidden = false;
  }
}

// actualizarCamposPorTipo muestra/oculta y (des)marca required en los campos
// de persona según el tipo elegido, y resincroniza la dirección de todas las
// filas de elementos ya creadas.
function actualizarCamposPorTipo() {
  const tipo = tipoActual();
  const mostrarEntrega = tipo === 'entrega' || tipo === 'ambos';

  document.body.classList.toggle('modo-ambos', tipo === 'ambos');
  // QuienEntrega conserva su valor (por defecto "Oficina de Sistemas") al
  // ocultarse: no tiene sentido pedirlo de nuevo cada vez que se vuelve a
  // mostrar.
  labelQuienEntrega.hidden = !mostrarEntrega;
  campoQuienEntrega.required = mostrarEntrega;

  textoAreaDepartamento.textContent = TEXTO_AREA_POR_TIPO[tipo];

  // Misma condición que aplicarDireccionEnFila: si el tipo fuerza una
  // única dirección, la columna de la cabecera se oculta igual que el
  // selector de cada fila, para que ambas sigan alineadas.
  headerDireccion.hidden = Boolean(direccionForzada(tipo));

  elementosContainer.querySelectorAll('.elemento-row').forEach((fila) => {
    aplicarDireccionEnFila(fila, tipo);
  });
}

tipoSelect.addEventListener('change', actualizarCamposPorTipo);
docPicker.addEventListener('change', actualizarDocumento);

// crearFilaElemento arma una fila nueva. Si se le pasan datosIniciales
// (usado al duplicar), prellena descripción/nroInventario/observación/
// dirección con esos valores. Devuelve la fila creada.
function crearFilaElemento(datosIniciales) {
  const fila = document.createElement('div');
  fila.className = 'elemento-row';
  // Cada campo va dentro de un <label class="campo">: el texto del label es
  // el nombre accesible y en pantallas chicas (sin cabecera de tabla) se
  // muestra arriba del input.
  fila.innerHTML = `
    <span class="elemento-numero" aria-hidden="true"></span>
    <label class="campo campo--descripcion">
      <span class="campo-label">Descripción</span>
      <input type="text" class="elemento-descripcion" placeholder="Ej. Monitor Samsung 24&quot;" required>
    </label>
    <label class="campo">
      <span class="campo-label">Nro. Inventario</span>
      <input type="text" class="elemento-nroInventario" placeholder="00000000" inputmode="numeric" pattern="[0-9]*" maxlength="8">
    </label>
    <label class="campo">
      <span class="campo-label">Cantidad</span>
      <input type="number" class="elemento-cantidad" placeholder="1" min="1" step="1" inputmode="numeric" value="1" required>
    </label>
    <label class="campo">
      <span class="campo-label">Dirección</span>
      <select class="elemento-direccion">
        <option value="retira">Retira</option>
        <option value="entrega">Entrega</option>
      </select>
    </label>
    <label class="campo campo--observacion">
      <span class="campo-label">Observación</span>
      <input type="text" class="elemento-observacion" placeholder="Opcional">
    </label>
    <div class="elemento-acciones">
      <button type="button" class="icon-btn duplicar-elemento" aria-label="Duplicar elemento" title="Duplicar">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="9" y="9" width="13" height="13" rx="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
      </button>
      <button type="button" class="icon-btn icon-btn--danger quitar-elemento" aria-label="Quitar elemento" title="Quitar">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/></svg>
      </button>
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
  aplicarDireccionEnFila(fila, tipoActual());

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

document.getElementById('agregar-elemento').addEventListener('click', () => (esRetiro() ? agregarElementoRetiro() : agregarFilaYEnfocar()));

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

// === Nota de retiro ===
// Sus elementos son solo una descripción cada uno y viven en su propio
// contenedor, así cambiar de documento no pisa lo cargado en el otro.
const retiroContainer = document.getElementById('retiro-container');

function crearElementoRetiro() {
  const fila = document.createElement('div');
  fila.className = 'elemento-row elemento-row--simple';
  fila.innerHTML = `
    <span class="elemento-numero" aria-hidden="true"></span>
    <input type="text" class="elemento-descripcion" placeholder="Ej. Impresora marca Pantum modelo M6550NW nro. de serie CJ3X001670" aria-label="Descripción del elemento" required>
    <div class="elemento-acciones">
      <button type="button" class="icon-btn icon-btn--danger quitar-elemento" aria-label="Quitar elemento" title="Quitar">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 6h18M8 6V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2M19 6l-1 14a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2L5 6"/></svg>
      </button>
    </div>
  `;
  // La nota exige al menos un elemento: la última fila no se puede quitar.
  fila.querySelector('.quitar-elemento').addEventListener('click', () => {
    if (retiroContainer.children.length > 1) fila.remove();
  });
  retiroContainer.appendChild(fila);
  return fila;
}

function agregarElementoRetiro() {
  crearElementoRetiro().querySelector('input').focus();
}

// actualizarDocumento adapta la pantalla al documento elegido: muestra solo
// las secciones .solo-acta o .solo-retiro y deshabilita los fieldsets
// ocultos (un fieldset disabled no se valida ni se envía), actualiza
// títulos, botón y resumen, y para las actas aplica las reglas por tipo.
const tituloDoc = document.getElementById('doc-titulo');
const descDoc = document.getElementById('doc-desc');
const resumenDoc = document.getElementById('resumen-doc');
const resumenItems = document.getElementById('resumen-items');

function actualizarDocumento() {
  const retiro = esRetiro();
  document.querySelectorAll('.solo-acta, .solo-retiro').forEach((el) => {
    const oculto = el.classList.contains('solo-acta') ? retiro : !retiro;
    el.hidden = oculto;
    if (el.tagName === 'FIELDSET') el.disabled = oculto;
  });

  const opcion = docPicker.querySelector('input:checked');
  tituloDoc.textContent = opcion.dataset.titulo;
  descDoc.textContent = opcion.dataset.desc;
  resumenDoc.textContent = opcion.nextElementSibling.textContent;
  if (!generandoEnCurso) botonGenerar.textContent = textoBoton();

  // Permite volver al mismo documento con un link (ej. index.html#retiro).
  history.replaceState(null, '', `#${documentoActual()}`);

  if (!retiro) actualizarCamposPorTipo();
  actualizarResumen();
}

function actualizarResumen() {
  const n = (esRetiro() ? retiroContainer : elementosContainer).querySelectorAll('.elemento-row').length;
  resumenItems.textContent = `${n} ${n === 1 ? 'elemento' : 'elementos'}`;
}

new MutationObserver(actualizarResumen).observe(elementosContainer, { childList: true });
new MutationObserver(actualizarResumen).observe(retiroContainer, { childList: true });

const textoBoton = () => (esRetiro() ? 'Generar Nota' : 'Generar Acta');

crearFilaElemento();
crearElementoRetiro();

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
// derecha, con fade-in/translateY de entrada, y la retira sola a los 5s
// (tiempo para alcanzar a usar el link de descarga).
// Si se pasa enlaceDescarga, agrega un link breve dentro del toast: sirve
// de respaldo si el navegador bloqueó la descarga automática (no hay forma
// confiable de detectar el bloqueo, así que siempre se ofrece).
function mostrarToast(mensaje, tipo, enlaceDescarga) {
  const toast = document.createElement('div');
  toast.className = `toast toast-${tipo}`;
  toast.setAttribute('role', tipo === 'error' ? 'alert' : 'status');
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
  }, 5000);
}

const botonGenerar = form.querySelector('button[type="submit"]');
let generandoEnCurso = false;

// Documento inicial: el del hash si es válido (ej. #retiro), si no "acta".
const opcionDelHash = Array.from(form.elements.documento).find((r) => r.value === location.hash.slice(1));
if (opcionDelHash) opcionDelHash.checked = true;
actualizarDocumento();

// leerNotaRetiro arma el cuerpo que espera /api/notas-retiro/generar.
function leerNotaRetiro(fecha) {
  return {
    fecha,
    empresa: document.getElementById('empresa').value,
    elementos: Array.from(retiroContainer.querySelectorAll('.elemento-descripcion')).map((i) => i.value),
  };
}

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

  const areaDepartamento = selectAreaDepartamento.value === 'otros'
    ? campoAreaDepartamentoOtra.value
    : selectAreaDepartamento.value;

  const retiro = esRetiro();
  const url = retiro ? '/api/notas-retiro/generar' : '/api/actas/generar';
  const cuerpo = retiro ? leerNotaRetiro(fecha) : {
    fecha,
    quienEntrega: document.getElementById('quienEntrega').value,
    tipo: tipoActual(),
    areaDepartamento,
    elementos: leerElementos(),
  };

  generandoEnCurso = true;
  botonGenerar.disabled = true;
  botonGenerar.innerHTML = '<span class="spinner" aria-hidden="true"></span> Generando PDF...';
  form.setAttribute('aria-busy', 'true');

  try {
    const respuesta = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(cuerpo),
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
    botonGenerar.textContent = textoBoton();
    form.removeAttribute('aria-busy');
  }
});
