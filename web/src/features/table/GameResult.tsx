import { MsgType } from '../../protocol/types';
import type { GameSocket } from '../../transport/wsClient';
import { useAppStore } from '../../stores/appStore';
import { PlayedCards } from '../../shared/cards/PlayedCards';

export function GameResult({ socket }: { socket: GameSocket }) {
  const winnerName = useAppStore((state) => state.winnerName);
  const winnerIsLandlord = useAppStore((state) => state.winnerIsLandlord);
  const finalMultiplier = useAppStore((state) => state.finalMultiplier);
  const scores = useAppStore((state) => state.scores);
  const playerHands = useAppStore((state) => state.playerHands);

  return (
    <main className="result-screen">
      <section className="result-panel">
        <span className="result-badge">{winnerIsLandlord ? '地主获胜' : '农民获胜'}</span>
        <h1>{winnerName || '本局'} 获胜</h1>
        <p>最终倍数 x{finalMultiplier || 1}</p>
        <div className="score-list">
          {scores.map((score) => (
            <div className="score-row" key={score.player_id}>
              <span>{score.player_name}</span>
              <strong className={score.score >= 0 ? 'is-positive' : 'is-negative'}>{score.score >= 0 ? '+' : ''}{score.score}</strong>
            </div>
          ))}
        </div>
        <div className="remaining-hands">
          {playerHands.map((playerHand) => (
            <PlayedCards key={playerHand.player_id} cards={playerHand.cards} playerName={playerHand.player_name} compact />
          ))}
        </div>
        <div className="room-actions">
          <button className="primary-action" onClick={() => socket.send(MsgType.Ready)}>再来一局</button>
          <button className="secondary-action" onClick={() => { socket.send(MsgType.LeaveRoom); useAppStore.getState().leaveLocalRoom(); }}>返回大厅</button>
        </div>
      </section>
    </main>
  );
}
