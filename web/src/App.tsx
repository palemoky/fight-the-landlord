import { useEffect, useMemo } from 'react';
import { Lobby } from './features/lobby/Lobby';
import { GameTable } from './features/table/GameTable';
import { GameResult } from './features/table/GameResult';
import { createGameSocket, type GameSocket } from './transport/wsClient';
import { useAppStore } from './stores/appStore';
import { seedDemoState } from './stores/demoState';

export function App() {
  const socket = useMemo<GameSocket>(() => createGameSocket(), []);
  const phase = useAppStore((state) => state.phase);
  const connected = useAppStore((state) => state.connected);
  const error = useAppStore((state) => state.error);
  const maintenance = useAppStore((state) => state.maintenance);

  useEffect(() => {
    const demo = new URLSearchParams(window.location.search).get('demo');
    if (demo) {
      seedDemoState(demo);
      return;
    }
    socket.connect();
    return () => socket.close();
  }, [socket]);

  return (
    <div className="app-shell">
      {maintenance ? <div className="maintenance-banner">服务器维护中，暂时无法开始新对局</div> : null}
      {!connected && !new URLSearchParams(window.location.search).get('demo') ? (
        <ConnectionState error={error} />
      ) : null}
      {phase === 'game_over' ? (
        <GameResult socket={socket} />
      ) : phase === 'bidding' || phase === 'playing' ? (
        <GameTable socket={socket} />
      ) : (
        <Lobby socket={socket} />
      )}
    </div>
  );
}

function ConnectionState({ error }: { error: string }) {
  return (
    <div className="connection-toast" role="status">
      <span className="spinner" />
      <span>{error || '正在连接服务器...'}</span>
    </div>
  );
}
