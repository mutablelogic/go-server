
/** 
 * @function
 * @name setFavIcon 
 * @description Set the FavIcon of the page
 * @param {string} url - The URL for the icon
 */
export function setFavIcon(url) {
    var link = document.querySelector("link[rel~='icon']");
    if (!link) {
        link = document.createElement('link');
        link.rel = 'icon';
        document.head.appendChild(link);
    }
    link.href = url;
}
