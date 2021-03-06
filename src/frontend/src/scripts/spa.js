/* 
    SPA SCRIPT

    Author : IvanK Production
*/

import { newAjax, ajaxErr, ProgressBar } from './ajax.js';
import { sleep, fadeOut, queryParse, queryStringify, getMeta, setMeta, rewriteMetas } from './utils.js';

//  Extras Data
let dataExtras = null;

//  HistoryAPI: state
const url = new URL(String(window.location.href));
const hState  = {
	href:   url.pathname,
	params: queryParse(url.searchParams.toString()),
	title:  document.title,
	url:    url.href
};

//  Loads ajax page
export async function loadPage(strHref, params = {}, changeAddress = false, callback = null) {
	const progress = new ProgressBar(); progress.start();
	let res = await newAjax(strHref, params, 'text');

	if (res.error && res.error.error_type == 'aborted') { return; }

	if (res.error && res.error.error_type == 'client') {
		ajaxErr(res.error.error_code, res.error.error_desc); return;
	}

	if (res.error && res.error.error_type == 'server') {
		res = res.response;
	}

	const sliderWrapper = document.querySelector('.slider-section');
	if (window.slider && sliderWrapper) { window.slider.destroy(); }
	
	const elemActiveNavItem = document.querySelector('ul.mnav li a.nav-item-active');
	if (elemActiveNavItem) elemActiveNavItem.classList.remove('nav-item-active');

	const oParser        = new DOMParser();
	const oDoc           = oParser.parseFromString(res, 'text/html');
	const newContent     = oDoc.getElementById('content');
	const oldContent     = document.getElementById('content');
	const newAuthInfo    = oDoc.getElementById('user-auth-info');
	const oldAuthInfo    = document.getElementById('user-auth-info');
	const containerAuth  = oldAuthInfo.parentNode;
	const newBreadcrumbs = oDoc.getElementById('breadcrumbs');
	const oldBreadcrumbs = document.getElementById('breadcrumbs');
	const container      = oldContent.parentNode;
	document.title       = oDoc.title;

	rewriteMetas({
		docSource: oDoc,
		docDest:   document,
		metas: [
			'robots',
			'og:title', 'og:description', 'og:type', 'og:image', 'og:url', 'og:site_name', 'og:locale',
			'twitter:card', 'twitter:title', 'twitter:description', 'twitter:image'
		],
		withCanonical: true
	});

	const elemDataExtras = oDoc.getElementById('data-extras');
	if (elemDataExtras) dataExtras = JSON.parse(elemDataExtras.textContent);

	hState.href   = strHref;
	hState.params = params;
	hState.title  = oDoc.title;
	hState.url    = url.origin + strHref + queryStringify(params);

	if (changeAddress) window.history.pushState(hState, hState.title, hState.url);

	if (newAuthInfo) {
		if (oldAuthInfo) { containerAuth.replaceChild(newAuthInfo, oldAuthInfo); }
		else { containerAuth.append(newAuthInfo); }
	} else {
		if (oldAuthInfo) { oldAuthInfo.remove(); }
	}

	if (newBreadcrumbs) {
		if (oldBreadcrumbs) { container.replaceChild(newBreadcrumbs, oldBreadcrumbs); }
		else { container.insertBefore(newBreadcrumbs, container.firstChild); }
	} else {
		if (oldBreadcrumbs) { oldBreadcrumbs.remove(); }
	}

	fadeOut(oldContent).then(() => {
		sleep(110).then(() => {
			container.replaceChild(newContent, oldContent); window.onPageLoaded(dataExtras);

			window.scrollTo({ top: 0, behavior: 'smooth' });
			document.querySelectorAll('.subnav-container').forEach(elem => {
				elem.classList.remove("showed");
			});

			progress.finish();
		});
	});

	const scope = getMeta(oDoc, 'app:scope');
	setMeta(document, 'app:scope', scope);

	const itemActive = document.querySelector('ul.mnav li a[data-scope="' + scope + '"]');
	if (itemActive) itemActive.classList.add('nav-item-active');

	if (callback) callback();
}

//  HistoryAPI: replace state on load
export function init() {
	window.history.replaceState(hState, hState.title, hState.url);
};

//  HistoryAPI: when back or forward
export function popstate(oEvent) {
	loadPage(oEvent.state.href, oEvent.state.params);
};

//  Exports
export default {
	init, popstate, loadPage
}
