import { Component, OnInit, computed, inject, signal } from '@angular/core';
import { FormBuilder, ReactiveFormsModule, Validators } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { catchError, forkJoin, of } from 'rxjs';

import { AuthService } from '../../../core/auth/auth.service';
import {
  ActualizarUsuarioRequest,
  CrearUsuarioRequest,
  RolSistema,
  SucursalOpcion,
  UsuarioItem,
} from '../usuarios.models';
import { UsuariosService } from '../usuarios.service';

@Component({
  selector: 'app-usuarios',
  standalone: true,
  imports: [ReactiveFormsModule],
  templateUrl: './usuarios.html',
  styleUrl: './usuarios.scss',
})
export class UsuariosComponent implements OnInit {
  private readonly fb = inject(FormBuilder);
  private readonly auth = inject(AuthService);
  private readonly usuariosSvc = inject(UsuariosService);

  readonly esAdministrador = this.auth.esAdministrador;
  readonly esAdminSucursal = this.auth.esAdminSucursal;
  readonly puedeGestionar = this.auth.puedeAdministrarUsuarios;
  readonly usuarioActualId = computed(() => this.auth.usuario()?.id ?? null);

  readonly loading = signal(true);
  readonly saving = signal(false);
  readonly error = signal<string | null>(null);
  readonly modalError = signal<string | null>(null);
  readonly success = signal<string | null>(null);

  readonly usuarios = signal<UsuarioItem[]>([]);
  readonly total = signal(0);
  readonly page = signal(1);
  readonly pageSize = 20;

  readonly roles = signal<RolSistema[]>([]);
  readonly sucursales = signal<SucursalOpcion[]>([]);

  readonly sucursalesMap = computed(
    () => new Map(this.sucursales().map((s) => [s.id, s.nombre])),
  );

  // Filtros
  readonly filtroQ = signal('');
  readonly filtroSucursalId = signal<number | null>(null);
  readonly filtroActivo = signal<'' | 'true' | 'false'>('');

  // Modal crear/editar usuario completo
  readonly modalAbierto = signal(false);
  readonly editando = signal<UsuarioItem | null>(null);
  readonly sucursalesSeleccionadas = signal<number[]>([]);
  readonly rolesSeleccionados = signal<number[]>([]);
  readonly errorRoles = signal<string | null>(null);
  readonly errorSucursales = signal<string | null>(null);

  // Modal rápido de asignación de roles y permisos de sucursales
  readonly modalPermisosUsuario = signal<UsuarioItem | null>(null);
  readonly permisosRolIds = signal<number[]>([]);
  readonly permisosSucursales = signal<number[]>([]);
  readonly permisosError = signal<string | null>(null);

  readonly rolSeleccionadoEsAdmin = computed(() => {
    const rids = this.rolesSeleccionados();
    return this.roles().some(
      (r) => rids.includes(r.id) && r.codigo.toUpperCase() === 'ADMINISTRADOR',
    );
  });

  readonly permisosRolEsAdmin = computed(() => {
    const rids = this.permisosRolIds();
    return this.roles().some(
      (r) => rids.includes(r.id) && r.codigo.toUpperCase() === 'ADMINISTRADOR',
    );
  });

  // Modal confirmar eliminación
  readonly usuarioAEliminar = signal<UsuarioItem | null>(null);

  readonly totalPaginas = computed(() => Math.max(1, Math.ceil(this.total() / this.pageSize)));

  readonly form = this.fb.nonNullable.group({
    nombre_usuario: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(50)]],
    nombres: ['', [Validators.required, Validators.maxLength(100)]],
    apellidos: ['', [Validators.required, Validators.maxLength(100)]],
    email: ['', [Validators.required, Validators.email, Validators.maxLength(255)]],
    password: ['', [Validators.minLength(8), Validators.maxLength(72)]],
    activo: [true],
  });

  ngOnInit(): void {
    this.cargarCatalogosYUsuarios();
  }

  cargarCatalogosYUsuarios(): void {
    this.loading.set(true);
    this.error.set(null);

    forkJoin({
      roles: this.usuariosSvc.listarRoles().pipe(catchError(() => of([] as RolSistema[]))),
      sucursales: this.usuariosSvc
        .listarSucursales()
        .pipe(catchError(() => of([] as SucursalOpcion[]))),
    }).subscribe(({ roles, sucursales }) => {
      this.roles.set(roles);
      this.sucursales.set(sucursales);
      this.cargarUsuarios();
    });
  }

  cargarUsuarios(): void {
    this.loading.set(true);
    this.error.set(null);

    const activoStr = this.filtroActivo();
    const activo = activoStr === '' ? undefined : activoStr === 'true';

    this.usuariosSvc
      .listar({
        q: this.filtroQ() || undefined,
        sucursal_id: this.filtroSucursalId() ?? undefined,
        activo,
        page: this.page(),
        page_size: this.pageSize,
      })
      .subscribe({
        next: (res) => {
          this.usuarios.set(res.data ?? []);
          this.total.set(res.total ?? 0);
          this.loading.set(false);
        },
        error: (err) => {
          this.error.set(this.extraerMensajeError(err, 'No se pudo cargar el listado de usuarios.'));
          this.loading.set(false);
        },
      });
  }

  onBuscar(valor: string): void {
    this.filtroQ.set(valor);
    this.page.set(1);
    this.cargarUsuarios();
  }

  onFiltroSucursal(valor: string): void {
    this.filtroSucursalId.set(valor ? Number(valor) : null);
    this.page.set(1);
    this.cargarUsuarios();
  }

  onFiltroActivo(valor: string): void {
    this.filtroActivo.set(valor as '' | 'true' | 'false');
    this.page.set(1);
    this.cargarUsuarios();
  }

  limpiarFiltros(): void {
    this.filtroQ.set('');
    this.filtroSucursalId.set(null);
    this.filtroActivo.set('');
    this.page.set(1);
    this.cargarUsuarios();
  }

  cambiarPagina(delta: number): void {
    const nueva = this.page() + delta;
    if (nueva < 1 || nueva > this.totalPaginas()) return;
    this.page.set(nueva);
    this.cargarUsuarios();
  }

  // ---------- Modal Crear / Editar usuario ----------

  puedeModificarUsuario(u: UsuarioItem): boolean {
    if (this.esAdministrador()) return true;
    if (this.esAdminSucursal()) return !u.es_administrador;
    return false;
  }

  abrirCrear(): void {
    if (!this.puedeGestionar()) return;
    this.editando.set(null);
    this.error.set(null);
    this.modalError.set(null);
    this.success.set(null);
    this.errorRoles.set(null);
    this.errorSucursales.set(null);

    const rolDefault =
      this.roles().find((r) => r.codigo.toUpperCase() === 'USUARIO')?.id ??
      this.roles()[0]?.id ??
      null;

    this.form.reset({
      nombre_usuario: '',
      nombres: '',
      apellidos: '',
      email: '',
      password: '',
      activo: true,
    });
    this.form.controls.password.setValidators([
      Validators.required,
      Validators.minLength(8),
      Validators.maxLength(72),
    ]);
    this.form.controls.password.updateValueAndValidity();

    this.rolesSeleccionados.set(rolDefault ? [rolDefault] : []);
    // Si solo tiene 1 sucursal disponible (típico de Administrador de sucursal), preseleccionarla
    const sucs = this.sucursales();
    this.sucursalesSeleccionadas.set(sucs.length === 1 ? [sucs[0].id] : []);
    this.modalAbierto.set(true);
  }

  abrirEditar(u: UsuarioItem): void {
    if (!this.puedeModificarUsuario(u)) return;
    this.editando.set(u);
    this.error.set(null);
    this.modalError.set(null);
    this.success.set(null);
    this.errorRoles.set(null);
    this.errorSucursales.set(null);

    const rolIds = (u.roles ?? []).map((r) => r.id);

    this.form.reset({
      nombre_usuario: u.nombre_usuario,
      nombres: u.nombres,
      apellidos: u.apellidos,
      email: u.email,
      password: '',
      activo: u.activo,
    });
    this.form.controls.password.setValidators([
      Validators.minLength(8),
      Validators.maxLength(72),
    ]);
    this.form.controls.password.updateValueAndValidity();

    this.rolesSeleccionados.set(rolIds);
    this.sucursalesSeleccionadas.set([...(u.sucursal_ids ?? [])]);
    this.modalAbierto.set(true);
  }

  cerrarModal(): void {
    this.modalAbierto.set(false);
    this.editando.set(null);
    this.modalError.set(null);
    this.errorRoles.set(null);
    this.errorSucursales.set(null);
  }

  estaRolSeleccionado(id: number): boolean {
    return this.rolesSeleccionados().includes(id);
  }

  toggleRol(id: number): void {
    this.errorRoles.set(null);
    this.errorSucursales.set(null);
    this.rolesSeleccionados.update((actual) =>
      actual.includes(id) ? actual.filter((x) => x !== id) : [...actual, id],
    );
  }

  estaSucursalSeleccionada(id: number): boolean {
    return this.sucursalesSeleccionadas().includes(id);
  }

  toggleSucursal(id: number): void {
    this.errorSucursales.set(null);
    this.sucursalesSeleccionadas.update((actual) =>
      actual.includes(id) ? actual.filter((x) => x !== id) : [...actual, id],
    );
  }

  seleccionarTodasSucursales(): void {
    this.errorSucursales.set(null);
    this.sucursalesSeleccionadas.set(this.sucursales().map((s) => s.id));
  }

  quitarTodasSucursales(): void {
    this.sucursalesSeleccionadas.set([]);
  }

  campoInvalido(nombre: keyof typeof this.form.controls): boolean {
    const ctrl = this.form.controls[nombre];
    return ctrl.invalid && (ctrl.dirty || ctrl.touched);
  }

  guardar(): void {
    if (!this.puedeGestionar()) return;
    this.modalError.set(null);

    if (this.form.invalid) {
      this.form.markAllAsTouched();
      return;
    }

    const v = this.form.getRawValue();
    const rolIds = this.rolesSeleccionados();
    if (rolIds.length === 0) {
      this.errorRoles.set('Debes asignar al menos un rol de sistema al usuario.');
      return;
    }

    const esAdmin = this.rolSeleccionadoEsAdmin();
    const sids = this.sucursalesSeleccionadas();

    if (!esAdmin && sids.length === 0) {
      this.errorSucursales.set(
        'Debes seleccionar al menos una sucursal para usuarios que no son administradores.',
      );
      return;
    }

    this.saving.set(true);
    const actual = this.editando();

    if (actual) {
      const payload: ActualizarUsuarioRequest = {
        rol_ids: rolIds,
        nombre_usuario: v.nombre_usuario.trim(),
        nombres: v.nombres.trim(),
        apellidos: v.apellidos.trim(),
        email: v.email.trim(),
        activo: v.activo,
      };
      if (v.password.trim()) {
        payload.password = v.password;
      }
      if (!esAdmin) {
        payload.sucursal_ids = sids;
      }

      this.usuariosSvc.actualizar(actual.id, payload).subscribe({
        next: () => {
          this.saving.set(false);
          this.cerrarModal();
          this.success.set('Usuario actualizado correctamente.');
          this.cargarUsuarios();
        },
        error: (err) => {
          this.saving.set(false);
          this.modalError.set(this.extraerMensajeError(err, 'No se pudo actualizar el usuario.'));
        },
      });
    } else {
      const payload: CrearUsuarioRequest = {
        rol_ids: rolIds,
        nombre_usuario: v.nombre_usuario.trim(),
        nombres: v.nombres.trim(),
        apellidos: v.apellidos.trim(),
        email: v.email.trim(),
        password: v.password,
        activo: v.activo,
      };
      if (!esAdmin) {
        payload.sucursal_ids = sids;
      }

      this.usuariosSvc.crear(payload).subscribe({
        next: () => {
          this.saving.set(false);
          this.cerrarModal();
          this.success.set('Usuario creado correctamente.');
          this.cargarUsuarios();
        },
        error: (err) => {
          this.saving.set(false);
          this.modalError.set(this.extraerMensajeError(err, 'No se pudo crear el usuario.'));
        },
      });
    }
  }

  // ---------- Modal rápido de Roles y Sucursales ----------

  abrirPermisos(u: UsuarioItem): void {
    if (!this.puedeModificarUsuario(u)) return;
    this.modalPermisosUsuario.set(u);
    this.permisosRolIds.set((u.roles ?? []).map((r) => r.id));
    this.permisosSucursales.set([...(u.sucursal_ids ?? [])]);
    this.permisosError.set(null);
    this.success.set(null);
    this.error.set(null);
  }

  cerrarPermisos(): void {
    this.modalPermisosUsuario.set(null);
    this.permisosError.set(null);
  }

  estaPermisoRol(id: number): boolean {
    return this.permisosRolIds().includes(id);
  }

  togglePermisoRol(id: number): void {
    this.permisosError.set(null);
    this.permisosRolIds.update((actual) =>
      actual.includes(id) ? actual.filter((x) => x !== id) : [...actual, id],
    );
  }

  estaPermisoSucursal(id: number): boolean {
    return this.permisosSucursales().includes(id);
  }

  togglePermisoSucursal(id: number): void {
    this.permisosError.set(null);
    this.permisosSucursales.update((actual) =>
      actual.includes(id) ? actual.filter((x) => x !== id) : [...actual, id],
    );
  }

  seleccionarTodasPermisos(): void {
    this.permisosError.set(null);
    this.permisosSucursales.set(this.sucursales().map((s) => s.id));
  }

  quitarTodasPermisos(): void {
    this.permisosSucursales.set([]);
  }

  guardarPermisos(): void {
    const u = this.modalPermisosUsuario();
    const rolIds = this.permisosRolIds();
    if (!u || !this.puedeModificarUsuario(u)) return;

    if (rolIds.length === 0) {
      this.permisosError.set('Debes asignar al menos un rol de sistema al usuario.');
      return;
    }

    const esAdmin = this.permisosRolEsAdmin();
    const sids = this.permisosSucursales();
    if (!esAdmin && sids.length === 0) {
      this.permisosError.set(
        'Debes otorgar acceso al menos a una sucursal para usuarios que no son administradores.',
      );
      return;
    }

    this.saving.set(true);
    this.permisosError.set(null);

    const payload: ActualizarUsuarioRequest = {
      rol_ids: rolIds,
    };
    if (!esAdmin) {
      payload.sucursal_ids = sids;
    }

    this.usuariosSvc.actualizar(u.id, payload).subscribe({
      next: () => {
        this.saving.set(false);
        this.cerrarPermisos();
        this.success.set(
          `Permisos de rol y sucursales actualizados para "${u.nombres} ${u.apellidos}".`,
        );
        this.cargarUsuarios();
      },
      error: (err) => {
        this.saving.set(false);
        this.permisosError.set(
          this.extraerMensajeError(err, 'No se pudieron actualizar los permisos del usuario.'),
        );
      },
    });
  }

  // ---------- Eliminar usuario ----------

  confirmarEliminar(u: UsuarioItem): void {
    if (!this.puedeModificarUsuario(u)) return;
    this.usuarioAEliminar.set(u);
  }

  cancelarEliminar(): void {
    this.usuarioAEliminar.set(null);
  }

  ejecutarEliminar(): void {
    const u = this.usuarioAEliminar();
    if (!u || !this.puedeModificarUsuario(u)) return;
    this.saving.set(true);
    this.usuariosSvc.eliminar(u.id).subscribe({
      next: () => {
        this.saving.set(false);
        this.usuarioAEliminar.set(null);
        this.success.set(`Usuario "${u.nombre_usuario}" eliminado correctamente.`);
        this.cargarUsuarios();
      },
      error: (err) => {
        this.saving.set(false);
        this.usuarioAEliminar.set(null);
        this.error.set(this.extraerMensajeError(err, 'No se pudo eliminar el usuario.'));
      },
    });
  }

  nombreSucursal(id: number): string {
    return this.sucursalesMap().get(id) ?? `Sucursal #${id}`;
  }

  private extraerMensajeError(err: unknown, fallback: string): string {
    if (err instanceof HttpErrorResponse) {
      const apiMsg = err.error?.error?.message;
      if (typeof apiMsg === 'string' && apiMsg.trim()) return apiMsg;
    }
    return fallback;
  }
}
