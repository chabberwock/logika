
export{}

declare global {
    interface Window {
        API: any;
    }
}

window.API = window.API || {};

window.API.greet = async function () {

};