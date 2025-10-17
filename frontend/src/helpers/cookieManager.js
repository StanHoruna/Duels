import Cookies from 'js-cookie';

class CookieManager {
  static setItem(name, value, expires = Date.now() + 30 * 24 * 60 * 60 * 1000) {
    Cookies.set('duels::' + name, value,
      {
        expires: new Date(expires),
        secure: true,
        sameSite: 'strict',
      }
    );
  }

  static getItem(name) {
    return Cookies.get('duels::' + name);
  }

  static removeItem(name) {
    Cookies.remove('duels::' + name);
  }
}

export default CookieManager;

