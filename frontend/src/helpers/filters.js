export function toFix(value, fix = 2) {
  if (isNaN(+value)) return 0;
  else return Math.round(+value * 10 ** fix) / 10 ** fix;
}

export const getFile = (item) => {
  if (item?.includes('user_uploads')) {
    return import.meta.env.VITE_API_URL + '/media' + item;
  } else {
    return import.meta.env.VITE_API_URL + '/static/' + item;
  }
};
