/*
 * SPCase graphic grammar — one small family of flat vector forms with a
 * product meaning, reused contextually across the public composition:
 *
 *   case  = an open ring (an unresolved problem; the gap closes at the final)
 *   team  = independent elements — disc, block, half-disc
 *   story = the elements scatter → converge → assemble into one resolved mark
 *
 * Primitives (`Disc`, `Block`, `HalfDisc`, `CaseArc`) are SVG fragments meant
 * to be composed inside a parent <svg>. Composed figures (`BrandMark`,
 * `ResolvedMark`) render their own <svg>. Everything is decorative: parent
 * SVGs must carry aria-hidden. Composed figures are tinted through the
 * --gm-* custom properties so one mark works on paper, sun and accent fields.
 */

interface DiscProps {
  cx: number;
  cy: number;
  r: number;
  fill: string;
  className?: string;
}

export function Disc({ cx, cy, r, fill, className }: DiscProps) {
  return <circle cx={cx} cy={cy} r={r} fill={fill} className={className} />;
}

interface BlockProps {
  cx: number;
  cy: number;
  size: number;
  fill: string;
  rotate?: number;
  className?: string;
}

export function Block({ cx, cy, size, fill, rotate = 0, className }: BlockProps) {
  return (
    <rect
      x={cx - size / 2}
      y={cy - size / 2}
      width={size}
      height={size}
      fill={fill}
      transform={rotate ? `rotate(${rotate} ${cx} ${cy})` : undefined}
      className={className}
    />
  );
}

interface HalfDiscProps {
  cx: number;
  cy: number;
  r: number;
  fill: string;
  /** Dome points up at 0°; rotation is about the center. */
  rotate?: number;
  className?: string;
}

export function HalfDisc({ cx, cy, r, fill, rotate = 0, className }: HalfDiscProps) {
  return (
    <path
      d={`M ${cx - r} ${cy} A ${r} ${r} 0 0 0 ${cx + r} ${cy} Z`}
      fill={fill}
      transform={rotate ? `rotate(${rotate} ${cx} ${cy})` : undefined}
      className={className}
    />
  );
}

interface CaseArcProps {
  cx: number;
  cy: number;
  r: number;
  width: number;
  stroke: string;
  /** Opening of the ring in degrees; 0 draws a closed circle. */
  gap?: number;
  /** Rotation of the gap center, degrees clockwise from 3 o'clock. */
  rotate?: number;
  className?: string;
}

function polar(cx: number, cy: number, r: number, deg: number): string {
  const a = (deg * Math.PI) / 180;
  return `${(cx + r * Math.cos(a)).toFixed(1)} ${(cy + r * Math.sin(a)).toFixed(1)}`;
}

export function CaseArc({
  cx,
  cy,
  r,
  width,
  stroke,
  gap = 70,
  rotate = 0,
  className,
}: CaseArcProps) {
  if (gap <= 0) {
    return (
      <circle
        cx={cx}
        cy={cy}
        r={r}
        fill="none"
        stroke={stroke}
        strokeWidth={width}
        className={className}
      />
    );
  }
  const start = rotate + gap / 2;
  const end = rotate + 360 - gap / 2;
  return (
    <path
      d={`M ${polar(cx, cy, r, start)} A ${r} ${r} 0 1 1 ${polar(cx, cy, r, end)}`}
      fill="none"
      stroke={stroke}
      strokeWidth={width}
      strokeLinecap="round"
      className={className}
    />
  );
}

/*
 * Brand mark: two team elements mid-assembly — the smallest telling of the
 * grammar. Used in the header and the mobile menu.
 */
export function BrandMark({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 28 22" aria-hidden="true" focusable="false">
      <rect x="1" y="8" width="13" height="13" fill="var(--gm-block, var(--color-accent))" />
      <circle cx="19.5" cy="8.5" r="7.5" fill="var(--gm-disc, var(--color-ink))" />
    </svg>
  );
}

/*
 * Resolved mark: the end state of the story — the case ring closed around
 * one assembled composition (block + half-disc + disc). Used where the
 * narrative completes: the final stage vignette and the footer.
 */
export function ResolvedMark({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 120 120" aria-hidden="true" focusable="false">
      <circle
        cx="60"
        cy="60"
        r="52"
        fill="none"
        stroke="var(--gm-ring, var(--color-ink))"
        strokeWidth="4"
      />
      <path d="M 30 74 A 30 30 0 0 0 90 74 Z" fill="var(--gm-dome, var(--color-accent))" />
      <rect x="34" y="74" width="52" height="18" fill="var(--gm-block, var(--color-ink))" />
      <circle cx="60" cy="52" r="12" fill="var(--gm-disc, var(--color-paper))" />
    </svg>
  );
}
