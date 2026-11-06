import type { CardInfo } from '../../protocol/types';
import { Card } from './Card';

interface PlayedCardsProps {
  cards: CardInfo[];
  handType?: string;
  playerName?: string;
  compact?: boolean;
}

export function PlayedCards({ cards, handType, playerName, compact = false }: PlayedCardsProps) {
  if (!cards.length) {
    return <div className="played-cards played-cards--empty">等待出牌</div>;
  }

  return (
    <div className={`played-cards ${compact ? 'played-cards--compact' : ''}`}>
      <div className="played-cards__meta">
        {playerName ? <span>{playerName}</span> : null}
        {handType ? <strong>{handType}</strong> : null}
      </div>
      <div className="played-cards__row">
        {cards.map((card, index) => (
          <Card key={`${card.suit}_${card.rank}_${index}`} card={card} size="played" />
        ))}
      </div>
    </div>
  );
}
