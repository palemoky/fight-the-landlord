import type { CSSProperties } from 'react';
import type { CardInfo } from '../../protocol/types';
import { normalizeCard } from './cardModel';

interface CardProps {
  card: CardInfo;
  selected?: boolean;
  size?: 'hand' | 'played' | 'mini' | 'action';
  style?: CSSProperties;
  onClick?: () => void;
}

export function Card({ card, selected = false, size = 'hand', style, onClick }: CardProps) {
  const face = normalizeCard(card);

  return (
    <button
      className={[
        'card',
        `card--${size}`,
        `card--${face.color}`,
        `card--${face.suit}`,
        selected ? 'is-selected' : '',
        face.isJoker ? 'is-joker' : ''
      ].filter(Boolean).join(' ')}
      type="button"
      style={style}
      onClick={onClick}
      aria-pressed={selected}
      aria-label={face.label}
    >
      {face.isJoker ? (
        <>
          <span className="card__joker-word">JOKER</span>
          <span className="card__joker-star">{face.suitSymbol}</span>
        </>
      ) : (
        <>
          <span className="card__index card__index--top">
            <b>{face.rankLabel}</b>
            <i>{face.suitSymbol}</i>
          </span>
          <span className="card__pip" aria-hidden="true">{face.suitSymbol}</span>
          <span className="card__index card__index--bottom" aria-hidden="true">
            <b>{face.rankLabel}</b>
            <i>{face.suitSymbol}</i>
          </span>
        </>
      )}
    </button>
  );
}
