import dayjs from 'dayjs';
import utc from 'dayjs/plugin/utc';
import timezone from 'dayjs/plugin/timezone';

export function getBuildTime() {
  dayjs.extend(utc);
  dayjs.extend(timezone);

  const buildTime = dayjs.tz(Date.now(), 'Europe/Rome').format('YYYY-MM-DD HH:mm:ss');

  return buildTime;
}
