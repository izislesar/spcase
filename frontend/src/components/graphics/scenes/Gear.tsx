/*
 * Wide gear motif used by several scenes. Teeth are rotated rounded rects
 * around a solid body — a deliberately simple, mechanical silhouette.
 * Rendered in flat fills only; decorative by contract (parents aria-hide
 * the whole scene).
 */
const ANGLES = [0, 45, 90, 135, 180, 225, 270, 315] as const;

export function Gear({
  cx,
  cy,
  r,
  fill,
  hubFill,
}: {
  cx: number;
  cy: number;
  r: number;
  fill: string;
  hubFill: string;
}) {
  const toothW = r * 0.36;
  const toothH = r * 0.44;
  return (
    <g>
      {ANGLES.map((angle) => (
        <rect
          key={angle}
          x={cx - toothW / 2}
          y={cy - r - toothH * 0.55}
          width={toothW}
          height={toothH}
          rx={toothW * 0.3}
          fill={fill}
          transform={`rotate(${angle} ${cx} ${cy})`}
        />
      ))}
      <circle cx={cx} cy={cy} r={r} fill={fill} />
      <circle cx={cx} cy={cy} r={r * 0.34} fill={hubFill} />
    </g>
  );
}
