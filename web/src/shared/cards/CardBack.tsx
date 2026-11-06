export function CardBack({ count = 1 }: { count?: number }) {
  return (
    <div className="card-back-stack" aria-label={`剩余 ${count} 张`}>
      {Array.from({ length: Math.min(count, 17) }, (_, index) => (
        <span key={index} className="card-back" />
      ))}
      <strong>{count}</strong>
    </div>
  );
}
