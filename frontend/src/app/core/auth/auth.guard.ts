import { CanActivateFn, Router } from '@angular/router';
import { inject } from '@angular/core';

import { AuthService } from './auth.service';

export const authGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  if (auth.isAuthenticated()) {
    return true;
  }
  return inject(Router).createUrlTree(['/login']);
};

export const guestGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  if (!auth.isAuthenticated()) {
    return true;
  }
  return inject(Router).createUrlTree(['/dashboard']);
};

export const adminGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  if (auth.isAuthenticated() && auth.esAdministrador()) {
    return true;
  }
  return inject(Router).createUrlTree(['/dashboard']);
};

export const adminOrBranchAdminGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  if (auth.isAuthenticated() && auth.puedeAdministrarUsuarios()) {
    return true;
  }
  return inject(Router).createUrlTree(['/dashboard']);
};

function rutaFallbackPorRol(auth: AuthService, router: Router) {
  if (auth.puedeAccederCRM()) return router.createUrlTree(['/crm/leads']);
  if (auth.puedeAccederVentas()) return router.createUrlTree(['/ventas']);
  if (auth.puedeAccederCompras()) return router.createUrlTree(['/compras']);
  return router.createUrlTree(['/dashboard']);
}

export const crmGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (!auth.isAuthenticated()) return router.createUrlTree(['/login']);
  return auth.puedeAccederCRM() ? true : rutaFallbackPorRol(auth, router);
};

export const ventasGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (!auth.isAuthenticated()) return router.createUrlTree(['/login']);
  return auth.puedeAccederVentas() ? true : rutaFallbackPorRol(auth, router);
};

export const comprasGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (!auth.isAuthenticated()) return router.createUrlTree(['/login']);
  return auth.puedeAccederCompras() ? true : rutaFallbackPorRol(auth, router);
};

export const tercerosGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (!auth.isAuthenticated()) return router.createUrlTree(['/login']);
  return auth.puedeAccederTerceros() ? true : rutaFallbackPorRol(auth, router);
};

export const operacionGuard: CanActivateFn = () => {
  const auth = inject(AuthService);
  const router = inject(Router);
  if (!auth.isAuthenticated()) return router.createUrlTree(['/login']);
  return auth.puedeAccederOperacion() ? true : rutaFallbackPorRol(auth, router);
};
