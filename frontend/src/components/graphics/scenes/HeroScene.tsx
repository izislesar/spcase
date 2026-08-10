import { Gear } from "./Gear";

/*
 * Hero scene: the case-solving machine as a wide editorial landscape.
 * Documents fly in from the left along a dashed path, the coral machine
 * works them over, and the championship pennant comes out on top. The
 * mustard field and the crank are deliberately cropped by the scene edges;
 * nothing sits centered inside a badge circle.
 *
 * Groups carry `scene-*` classes for the one-shot staggered entrance
 * (animation lives in Hero.module.css; reduced motion collapses it).
 */
export function HeroScene({ className }: { className?: string }) {
  return (
    <svg className={className} viewBox="0 0 960 640" aria-hidden="true" focusable="false">
      {/* backdrop fields, cropped at the scene edges */}
      <g className="scene-field">
        <circle cx="940" cy="140" r="330" fill="var(--color-mustard)" />
        <rect
          x="-60"
          y="390"
          width="200"
          height="200"
          fill="var(--color-turquoise)"
          transform="rotate(-14 40 490)"
        />
        <rect
          x="120"
          y="70"
          width="54"
          height="54"
          fill="var(--color-accent)"
          transform="rotate(12 147 97)"
        />
      </g>
      {/* soft ground shadow */}
      <ellipse cx="580" cy="566" rx="290" ry="18" fill="var(--color-ink)" fillOpacity="0.12" />
      {/* dashed travel path from the documents into the machine slot */}
      <path
        d="M 36 436 C 160 410 190 300 310 282 C 410 266 470 286 528 306"
        fill="none"
        stroke="var(--color-ink)"
        strokeWidth="5"
        strokeLinecap="round"
        strokeDasharray="2 14"
      />
      {/* incoming documents at different scales */}
      <g className="scene-docs">
        <g transform="rotate(-11 205 255)">
          <rect x="140" y="170" width="130" height="170" rx="8" fill="var(--color-surface)" />
          <path
            d="M 162 206 H 248 M 162 234 H 248 M 162 262 H 224"
            stroke="var(--color-ink)"
            strokeWidth="8"
            strokeLinecap="round"
          />
        </g>
        <g transform="rotate(8 325 370)">
          <rect x="280" y="310" width="104" height="130" rx="7" fill="var(--color-surface)" />
          <path
            d="M 296 344 H 368 M 296 368 H 368 M 296 392 H 344"
            stroke="var(--color-turquoise)"
            strokeWidth="8"
            strokeLinecap="round"
          />
        </g>
      </g>
      {/* the machine */}
      <g className="scene-machine">
        <rect x="540" y="520" width="34" height="34" rx="8" fill="var(--color-ink)" />
        <rect x="776" y="520" width="34" height="34" rx="8" fill="var(--color-ink)" />
        <rect x="496" y="286" width="340" height="244" rx="30" fill="var(--color-accent)" />
        <rect x="560" y="256" width="216" height="38" rx="19" fill="var(--color-ink)" />
        <circle cx="650" cy="410" r="66" fill="var(--color-surface)" />
        <Gear cx={650} cy={410} r={30} fill="var(--color-ink)" hubFill="var(--color-mustard)" />
        {/* crank, cropped by the right scene edge */}
        <circle cx="836" cy="400" r="26" fill="var(--color-ink)" />
        <rect x="836" y="390" width="76" height="18" rx="9" fill="var(--color-ink)" />
        <circle cx="918" cy="399" r="15" fill="var(--color-mustard)" />
      </g>
      {/* result: the championship pennant */}
      <g className="scene-flag">
        <rect x="646" y="80" width="14" height="178" rx="7" fill="var(--color-ink)" />
        <circle cx="653" cy="72" r="15" fill="var(--color-turquoise)" />
        <path d="M 660 92 L 872 142 L 660 194 Z" fill="var(--color-accent)" />
      </g>
    </svg>
  );
}
