export const ssr = false;
export const prerender = false;

export const load = () => {
	if (typeof window !== 'undefined' && !localStorage.getItem('access_token')) {
		throw new Response(null, { status: 303, headers: { location: '/login' } });
	}
};
