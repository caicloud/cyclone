import moment from 'moment';

export function FormatTime(time, formatStr = 'YYYY-MM-DD HH:mm:ss') {
  return moment(time).format(formatStr);
}
