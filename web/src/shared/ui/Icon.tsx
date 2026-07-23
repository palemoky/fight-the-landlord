interface IconProps {
  name: 'play' | 'room' | 'bot' | 'rank' | 'stats' | 'chat' | 'counter' | 'history' | 'rules' | 'close';
}

export function Icon({ name }: IconProps) {
  const path = icons[name];
  return (
    <svg className="icon" viewBox="0 0 24 24" aria-hidden="true">
      <path d={path} />
    </svg>
  );
}

const icons: Record<IconProps['name'], string> = {
  play: 'M8 5v14l11-7-11-7Z',
  room: 'M4 5h16v14H4V5Zm3 3v8h10V8H7Z',
  bot: 'M7 8h10a3 3 0 0 1 3 3v5a3 3 0 0 1-3 3H7a3 3 0 0 1-3-3v-5a3 3 0 0 1 3-3Zm2-4h6v2H9V4Zm0 8h2v2H9v-2Zm4 0h2v2h-2v-2Z',
  rank: 'M5 20V9h4v11H5Zm5 0V4h4v16h-4Zm5 0v-7h4v7h-4Z',
  stats: 'M5 4h14v4H5V4Zm0 6h9v4H5v-4Zm0 6h12v4H5v-4Z',
  chat: 'M4 5h16v11H8l-4 4V5Z',
  counter: 'M4 4h16v16H4V4Zm3 3v3h3V7H7Zm0 5v5h3v-5H7Zm5-5v3h5V7h-5Zm0 5v5h5v-5h-5Z',
  history: 'M12 5a7 7 0 1 1-6.3 4H3l3.5-4L10 9H7.9A5 5 0 1 0 12 7v5l4 2-.9 1.8-5.1-2.6V5h2Z',
  rules: 'M5 4h14v16H5V4Zm3 4h8V6H8v2Zm0 4h8v-2H8v2Zm0 4h5v-2H8v2Z',
  close: 'm6 7.4 1.4-1.4 4.6 4.6L16.6 6 18 7.4 13.4 12l4.6 4.6-1.4 1.4-4.6-4.6L7.4 18 6 16.6l4.6-4.6L6 7.4Z'
};
