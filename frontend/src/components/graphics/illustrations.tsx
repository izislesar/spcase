/*
 * SPCase illustration system — original flat vector scenes for the public
 * surface. One consistent style: thick simple silhouettes, flat fills from
 * the section palette (navy, coral, mustard, turquoise, white), small spark
 * accents and occasional dashed travel paths. No outlines-only clip art,
 * no gradients, no abstract identity geometry: each scene illustrates a
 * real piece of the product (machine solving the case, stages, schedule,
 * final, questions).
 *
 * Everything is decorative: callers must treat scenes as aria-hidden
 * (each component already renders its own <svg aria-hidden>).
 * Scenes are tinted exclusively through the global palette tokens so they
 * stay coherent on white, mustard, turquoise and navy fields.
 *
 * The hero scene plays a one-shot entrance assembly via the `scene-*`
 * group classes (defined in home.module.css); reduced motion collapses it
 * to the final static composition through the global rule in base.css.
 */

/* Four-point spark star; `scale` sets the size, center at (cx, cy). */
function Star4({
  cx,
  cy,
  scale = 1,
  fill,
}: {
  cx: number;
  cy: number;
  scale?: number;
  fill: string;
}) {
  return (
    <path
      d="M 0 -12 C 1.5 -3.5 3.5 -1.5 12 0 C 3.5 1.5 1.5 3.5 0 12 C -1.5 3.5 -3.5 1.5 -12 0 C -3.5 -1.5 -1.5 -3.5 0 -12 Z"
      fill={fill}
      transform={`translate(${cx} ${cy}) scale(${scale})`}
    />
  );
}

/* Small plus-mark spark. */
function Plus({
  cx,
  cy,
  size = 10,
  stroke,
  width = 5,
}: {
  cx: number;
  cy: number;
  size?: number;
  stroke: string;
  width?: number;
}) {
  return (
    <path
      d={`M ${cx - size} ${cy} H ${cx + size} M ${cx} ${cy - size} V ${cy + size}`}
      fill="none"
      stroke={stroke}
      strokeWidth={width}
      strokeLinecap="round"
    />
  );
}

/*
 * Brand mark: a small championship pennant — the simplest product symbol
 * (reaching the final), not an abstract identity shape. The pole takes
 * --mark-pole so the mark works on light and navy surfaces.
 */
export function Mark({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 24 24" aria-hidden="true" focusable="false">
      <rect
        x="4.5"
        y="2.5"
        width="3"
        height="19"
        rx="1.5"
        fill="var(--mark-pole, var(--color-ink))"
      />
      <path d="M 9 4 L 21.5 8.75 L 9 13.5 Z" fill="var(--color-accent)" />
    </svg>
  );
}

/*
 * Hero scene: the case-solving machine. Documents fly in along a dashed
 * path, the coral machine works them over with a crank, and a championship
 * pennant comes out on top. One large bespoke composition; groups carry
 * `scene-*` classes for the staggered entrance.
 */
export function CaseMachine({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 520 560" aria-hidden="true" focusable="false">
      {/* backdrop fields */}
      <g className="scene-backdrop">
        <circle cx="292" cy="252" r="204" fill="var(--color-mustard)" />
        <rect
          x="34"
          y="376"
          width="112"
          height="112"
          fill="var(--color-turquoise)"
          transform="rotate(-12 90 432)"
        />
        <circle cx="72" cy="86" r="26" fill="var(--color-accent)" />
      </g>
      {/* soft ground shadow */}
      <ellipse cx="330" cy="486" rx="158" ry="16" fill="var(--color-ink)" fillOpacity="0.12" />
      {/* dashed travel path from the documents into the machine slot */}
      <path
        d="M 40 316 C 96 296 118 240 172 224 C 216 211 244 232 268 250"
        fill="none"
        stroke="var(--color-ink)"
        strokeWidth="5"
        strokeLinecap="round"
        strokeDasharray="2 14"
      />
      {/* incoming documents */}
      <g className="scene-docs">
        <g transform="rotate(-12 118 200)">
          <rect x="73" y="142" width="90" height="116" rx="6" fill="var(--color-surface)" />
          <path
            d="M 88 168 H 148 M 88 190 H 148 M 88 212 H 128"
            stroke="var(--color-ink)"
            strokeWidth="7"
            strokeLinecap="round"
          />
        </g>
        <g transform="rotate(9 158 258)">
          <rect x="113" y="206" width="90" height="104" rx="6" fill="var(--color-surface)" />
          <path
            d="M 128 230 H 188 M 128 252 H 188 M 128 274 H 168"
            stroke="var(--color-turquoise)"
            strokeWidth="7"
            strokeLinecap="round"
          />
        </g>
      </g>
      {/* the machine */}
      <g className="scene-machine">
        <rect x="238" y="296" width="196" height="168" rx="18" fill="var(--color-accent)" />
        <rect x="272" y="268" width="128" height="30" rx="15" fill="var(--color-ink)" />
        <rect x="258" y="464" width="28" height="20" rx="6" fill="var(--color-ink)" />
        <rect x="386" y="464" width="28" height="20" rx="6" fill="var(--color-ink)" />
        {/* porthole with the work going on inside */}
        <circle cx="336" cy="382" r="40" fill="var(--color-surface)" />
        <path
          d="M 312 388 L 328 404 L 362 364"
          fill="none"
          stroke="var(--color-accent)"
          strokeWidth="10"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
        {/* crank */}
        <circle cx="434" cy="372" r="20" fill="var(--color-ink)" />
        <rect x="434" y="364" width="42" height="15" rx="7.5" fill="var(--color-ink)" />
        <circle cx="480" cy="371" r="11" fill="var(--color-mustard)" />
      </g>
      {/* result: the championship pennant */}
      <g className="scene-flag">
        <rect x="316" y="92" width="11" height="184" rx="5.5" fill="var(--color-ink)" />
        <circle cx="321.5" cy="88" r="12" fill="var(--color-turquoise)" />
        <path d="M 327 106 L 442 138 L 327 172 Z" fill="var(--color-accent)" />
      </g>
      {/* sparks */}
      <g className="scene-sparks">
        <Star4 cx={452} cy={210} fill="var(--color-ink)" />
        <Star4 cx={238} cy={128} scale={0.7} fill="var(--color-accent)" />
        <Plus cx={488} cy={286} stroke="var(--color-ink)" />
      </g>
    </svg>
  );
}

/* Stage 01 — registration: the application sheet with a pencil. */
export function StageSheet({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 240 200" aria-hidden="true" focusable="false">
      <circle cx="118" cy="102" r="84" fill="var(--color-mustard)" />
      <g transform="rotate(-6 115 100)">
        <rect x="68" y="30" width="104" height="134" rx="8" fill="var(--color-surface)" />
        <path
          d="M 84 58 H 156 M 84 80 H 156 M 84 102 H 138"
          stroke="var(--color-ink)"
          strokeWidth="8"
          strokeLinecap="round"
        />
        <path
          d="M 84 132 l 12 12 l 24 -26"
          fill="none"
          stroke="var(--color-accent)"
          strokeWidth="9"
          strokeLinecap="round"
          strokeLinejoin="round"
        />
      </g>
      <g transform="rotate(32 186 120)">
        <rect x="178" y="76" width="16" height="74" rx="6" fill="var(--color-accent)" />
        <path d="M 178 150 L 186 168 L 194 150 Z" fill="var(--color-ink)" />
      </g>
      <Star4 cx={50} cy={46} scale={0.7} fill="var(--color-accent)" />
    </svg>
  );
}

/* Stage 02 — the work: one strong idea over the task. */
export function StageBulb({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 240 200" aria-hidden="true" focusable="false">
      <rect
        x="44"
        y="24"
        width="152"
        height="152"
        fill="var(--color-turquoise)"
        transform="rotate(6 120 100)"
      />
      <path
        d="M 120 14 V 34 M 66 38 L 80 52 M 174 38 L 160 52"
        stroke="var(--color-ink)"
        strokeWidth="8"
        strokeLinecap="round"
      />
      <circle cx="120" cy="98" r="44" fill="var(--color-accent)" />
      <path
        d="M 100 84 A 26 26 0 0 1 126 72"
        fill="none"
        stroke="var(--color-surface)"
        strokeWidth="8"
        strokeLinecap="round"
      />
      <rect x="100" y="136" width="40" height="12" rx="6" fill="var(--color-ink)" />
      <rect x="104" y="152" width="32" height="12" rx="6" fill="var(--color-ink)" />
    </svg>
  );
}

/* Stage 03 — the final: the pennant is raised. */
export function StageFlag({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 240 200" aria-hidden="true" focusable="false">
      <circle cx="120" cy="102" r="82" fill="var(--color-mustard)" />
      <ellipse cx="118" cy="180" rx="62" ry="8" fill="var(--color-ink)" fillOpacity="0.12" />
      <rect x="106" y="30" width="11" height="146" rx="5.5" fill="var(--color-ink)" />
      <circle cx="111.5" cy="28" r="11" fill="var(--color-turquoise)" />
      <path d="M 117 42 L 210 68 L 117 96 Z" fill="var(--color-accent)" />
      <g transform="rotate(18 62 72)">
        <rect x="56" y="66" width="12" height="12" fill="var(--color-accent)" />
      </g>
      <g transform="rotate(-24 184 140)">
        <rect x="178" y="134" width="12" height="12" fill="var(--color-ink)" />
      </g>
      <Star4 cx={52} cy={128} scale={0.7} fill="var(--color-ink)" />
    </svg>
  );
}

/* Schedule: the marked calendar page. */
export function CalendarArt({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 240 200" aria-hidden="true" focusable="false">
      <circle cx="120" cy="100" r="84" fill="var(--color-turquoise)" />
      <rect x="84" y="22" width="9" height="22" rx="4.5" fill="var(--color-mustard)" />
      <rect x="147" y="22" width="9" height="22" rx="4.5" fill="var(--color-mustard)" />
      <rect x="48" y="34" width="144" height="128" rx="14" fill="var(--color-navy)" />
      <path
        d="M 48 48 A 14 14 0 0 1 62 34 H 178 A 14 14 0 0 1 192 48 V 66 H 48 Z"
        fill="var(--color-accent)"
      />
      <rect x="70" y="86" width="44" height="44" rx="9" fill="var(--color-surface)" />
      <path
        d="M 78 108 H 106 M 78 120 H 96"
        stroke="var(--color-ink)"
        strokeWidth="7"
        strokeLinecap="round"
      />
      <path
        d="M 128 106 l 13 13 l 27 -31"
        fill="none"
        stroke="var(--color-turquoise)"
        strokeWidth="10"
        strokeLinecap="round"
        strokeLinejoin="round"
      />
      <Star4 cx={196} cy={158} scale={0.75} fill="var(--color-accent)" />
    </svg>
  );
}

/* Final and the jury: the cup. */
export function CupArt({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 240 200" aria-hidden="true" focusable="false">
      <circle cx="120" cy="100" r="82" fill="var(--color-mustard)" />
      <path
        d="M 82 62 C 56 62 56 96 84 98 M 158 62 C 184 62 184 96 156 98"
        fill="none"
        stroke="var(--color-accent)"
        strokeWidth="11"
        strokeLinecap="round"
      />
      <path d="M 82 46 H 158 L 150 118 H 90 Z" fill="var(--color-accent)" />
      <Star4 cx={120} cy={80} scale={0.85} fill="var(--color-surface)" />
      <rect x="112" y="118" width="16" height="22" fill="var(--color-ink)" />
      <rect x="88" y="140" width="64" height="16" rx="8" fill="var(--color-ink)" />
      <Plus cx={52} cy={56} size={9} stroke="var(--color-ink)" />
      <Plus cx={192} cy={140} size={9} stroke="var(--color-turquoise)" />
    </svg>
  );
}

/* Questions: two overlapping conversation bubbles. */
export function ChatArt({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 240 200" aria-hidden="true" focusable="false">
      <circle cx="188" cy="44" r="22" fill="var(--color-accent)" />
      <path d="M 52 118 L 64 144 L 88 122 Z" fill="var(--color-navy)" />
      <rect x="36" y="42" width="118" height="76" rx="20" fill="var(--color-navy)" />
      <circle cx="72" cy="80" r="8" fill="var(--color-surface)" />
      <circle cx="96" cy="80" r="8" fill="var(--color-surface)" />
      <circle cx="120" cy="80" r="8" fill="var(--color-surface)" />
      <path d="M 188 136 L 176 162 L 152 140 Z" fill="var(--color-turquoise)" />
      <rect x="92" y="96" width="118" height="76" rx="20" fill="var(--color-turquoise)" />
      <circle cx="128" cy="134" r="8" fill="var(--color-accent)" />
      <circle cx="152" cy="134" r="8" fill="var(--color-accent)" />
      <circle cx="176" cy="134" r="8" fill="var(--color-accent)" />
    </svg>
  );
}

/*
 * Footer scene: the raised pennant at large scale with sparks — the closing
 * composition of the page. Simplified echo of the machine's result, not a
 * copy of the hero scene.
 */
export function FlagScene({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 320 240" aria-hidden="true" focusable="false">
      <circle cx="160" cy="126" r="104" fill="var(--color-mustard)" />
      <ellipse cx="160" cy="222" rx="84" ry="10" fill="var(--color-ink)" fillOpacity="0.14" />
      <rect x="140" y="36" width="13" height="182" rx="6.5" fill="var(--color-ink)" />
      <circle cx="146.5" cy="34" r="13" fill="var(--color-turquoise)" />
      <path d="M 153 52 L 272 86 L 153 122 Z" fill="var(--color-accent)" />
      <Star4 cx={64} cy={76} fill="var(--color-ink)" />
      <Star4 cx={266} cy={168} scale={0.7} fill="var(--color-surface)" />
      <Plus cx={84} cy={176} stroke="var(--color-accent)" />
    </svg>
  );
}
