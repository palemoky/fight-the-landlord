import { useMemo, useRef, useState, type CSSProperties } from 'react';
import type { CardInfo } from '../../protocol/types';
import { cardKey } from './cardModel';
import { Card } from './Card';

interface HandProps {
  cards: CardInfo[];
  selected: Set<string>;
  disabled?: boolean;
  onToggle: (key: string) => void;
  onRangeSelect: (keys: string[]) => void;
}

export interface HandCardLayout {
  card: CardInfo;
  key: string;
  index: number;
  groupIndex: number;
  groupSize: number;
  groupOffset: number;
  groupPosition: number;
  row: number;
  rowIndex: number;
  rowCount: number;
  singleX: number;
  compactX: number;
  rowX: number;
}

interface HandGroup {
  cards: CardInfo[];
  startIndex: number;
}

export function Hand({ cards, selected, disabled = false, onToggle, onRangeSelect }: HandProps) {
  const rootRef = useRef<HTMLDivElement | null>(null);
  const lastHitIndexRef = useRef<number | null>(null);
  const [dragStart, setDragStart] = useState<number | null>(null);
  const [dragMoved, setDragMoved] = useState(false);
  const [dragKeys, setDragKeys] = useState<Set<string>>(new Set());
  const layout = useMemo(() => buildHandLayout(cards), [cards]);
  const keys = useMemo(() => layout.map((item) => item.key), [layout]);

  function indexFromPoint(clientX: number, clientY: number): number {
    const target = document.elementFromPoint(clientX, clientY);
    return indexFromTarget(target);
  }

  function indexFromTarget(target: EventTarget | Element | null): number {
    const element = (target as HTMLElement | null)?.closest?.('[data-card-index]');
    return element ? Number((element as HTMLElement).dataset.cardIndex) : -1;
  }

  function applyRange(endIndex: number) {
    if (dragStart === null || endIndex < 0) return;
    lastHitIndexRef.current = endIndex;
    if (endIndex !== dragStart) setDragMoved(true);
    const min = Math.min(dragStart, endIndex);
    const max = Math.max(dragStart, endIndex);
    const rangedKeys = keys.slice(min, max + 1);
    setDragKeys(new Set(rangedKeys));
    onRangeSelect(Array.from(new Set([...selected, ...rangedKeys])));
  }

  function selectGroup(groupIndex: number) {
    const groupKeys = layout.filter((item) => item.groupIndex === groupIndex).map((item) => item.key);
    if (!groupKeys.length) return;
    const allSelected = groupKeys.every((key) => selected.has(key));
    const next = new Set(selected);
    for (const key of groupKeys) {
      if (allSelected) next.delete(key);
      else next.add(key);
    }
    onRangeSelect([...next]);
  }

  return (
    <div
      ref={rootRef}
      className={`hand ${disabled ? 'is-disabled' : ''} ${dragStart !== null ? 'is-dragging' : ''}`}
      style={{ '--hand-count': cards.length } as CSSProperties}
      role="listbox"
      aria-label="手牌"
      onPointerDown={(event) => {
        if (disabled) return;
        const index = indexFromTarget(event.target);
        if (index >= 0) {
          setDragStart(index);
          setDragMoved(false);
          lastHitIndexRef.current = index;
          setDragKeys(new Set([keys[index]]));
          rootRef.current?.setPointerCapture(event.pointerId);
        }
      }}
      onPointerMove={(event) => {
        if (disabled || dragStart === null) return;
        event.preventDefault();
        const index = indexFromPoint(event.clientX, event.clientY);
        applyRange(index);
      }}
      onPointerUp={(event) => {
        if (disabled) return;
        const index = indexFromPoint(event.clientX, event.clientY);
        const hitIndex = index >= 0 ? index : lastHitIndexRef.current;
        if (dragStart !== null && !dragMoved && hitIndex === dragStart) onToggle(keys[dragStart]);
        setDragStart(null);
        setDragMoved(false);
        lastHitIndexRef.current = null;
        setDragKeys(new Set());
        try {
          rootRef.current?.releasePointerCapture(event.pointerId);
        } catch {
          // Pointer capture may already be released by the browser.
        }
      }}
      onPointerCancel={() => {
        setDragStart(null);
        setDragMoved(false);
        lastHitIndexRef.current = null;
        setDragKeys(new Set());
      }}
    >
      {layout.map((item) => {
        const isSelected = selected.has(item.key) || dragKeys.has(item.key);
        return (
          <div
            key={`${item.key}_${item.index}`}
            className={`hand__slot ${isSelected ? 'is-selected' : ''}`}
            data-card-index={item.index}
            data-group-index={item.groupIndex}
            data-row={item.row}
            style={{
              '--i': item.index,
              '--group': item.groupIndex,
              '--group-offset': item.groupOffset,
              '--group-position': item.groupPosition,
              '--group-size': item.groupSize,
              '--row': item.row,
              '--row-index': item.rowIndex,
              '--row-count': item.rowCount,
              '--single-x': `${item.singleX}px`,
              '--compact-x': `${item.compactX}px`,
              '--row-x': `${item.rowX}px`,
              zIndex: item.index + (item.row === 1 ? 40 : 0)
            } as CSSProperties}
            role="option"
            aria-selected={isSelected}
            onDoubleClick={() => {
              if (!disabled) selectGroup(item.groupIndex);
            }}
          >
            <Card card={item.card} selected={isSelected} />
          </div>
        );
      })}
    </div>
  );
}

export function buildHandLayout(cards: CardInfo[]): HandCardLayout[] {
  const groups = buildGroups(cards);
  const splitAt = Math.ceil(groups.length / 2);
  const rowGroups = [groups.slice(0, splitAt), groups.slice(splitAt)];
  const singlePositions = buildCenteredPositions(groups, 32, 8);
  const compactPositions = buildCenteredPositions(groups, 25, 5);
  const rowPositions = new Map<number, number>();

  rowGroups.forEach((row) => {
    for (const [index, x] of buildCenteredPositions(row, 30, 7)) rowPositions.set(index, x);
  });

  const items: HandCardLayout[] = [];

  rowGroups.forEach((row, rowIndex) => {
    const rowCards = row.flatMap((group) => group.cards);
    let cursor = 0;
    row.forEach((group, groupPosition) => {
      const groupIndex = rowIndex === 0 ? groupPosition : splitAt + groupPosition;
      group.cards.forEach((card, groupOffset) => {
        const index = group.startIndex + groupOffset;
        items.push({
          card,
          key: cardKey(card),
          index,
          groupIndex,
          groupSize: group.cards.length,
          groupOffset,
          groupPosition,
          row: rowIndex,
          rowIndex: cursor,
          rowCount: rowCards.length,
          singleX: singlePositions.get(index) ?? 0,
          compactX: compactPositions.get(index) ?? 0,
          rowX: rowPositions.get(index) ?? 0
        });
        cursor += 1;
      });
    });
  });

  return items.sort((a, b) => a.index - b.index);
}

function buildGroups(cards: CardInfo[]): HandGroup[] {
  const groups: HandGroup[] = [];
  cards.forEach((card, index) => {
    const current = groups[groups.length - 1];
    if (!current || current.cards[0].rank !== card.rank) groups.push({ cards: [card], startIndex: index });
    else current.cards.push(card);
  });
  return groups;
}

function buildCenteredPositions(groups: HandGroup[], cardStep: number, groupGap: number): Map<number, number> {
  const positions = new Map<number, number>();
  let cursor = 0;

  groups.forEach((group, groupIndex) => {
    group.cards.forEach((_, groupOffset) => {
      positions.set(group.startIndex + groupOffset, cursor);
      cursor += cardStep;
    });
    if (groupIndex < groups.length - 1) cursor += groupGap;
  });

  const values = [...positions.values()];
  if (!values.length) return positions;
  const center = (Math.min(...values) + Math.max(...values)) / 2;
  for (const [index, x] of positions) positions.set(index, Math.round(x - center));
  return positions;
}
