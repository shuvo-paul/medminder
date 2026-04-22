import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';

export const ssr = false;
export const prerender = false;

export const load: LayoutLoad = () => {
	if (typeof window !== 'undefined' && !localStorage.getItem('access_token')) {
		throw redirect(302, '/login');
	}
};
