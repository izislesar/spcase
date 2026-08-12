# Frontend design direction

> **Status: approved stable visual constitution for Stage 4H and subsequent
> product-surface work unless superseded by human review.**
> Current human verdicts live in `visual-acceptance.md`; product/interaction
> semantics live in `experience-model.md`.

## North star

SPCase is a **dark spatial competition interface**.

It should feel like a real championship with a strong digital presence, not a
SaaS dashboard, a creative-agency portfolio, a Swiss-poster exercise or a
3D-showcase website.

Stage 4F established the correct neutral foundation: dark canvas, factual copy,
low decorative noise and much less motion. Stage 4G then established the
approved spatial/material vocabulary. Human review accepted that direction but
found the rendered composition **too tentative**: the hero artifact is small
and risks reading as a layered-card stack, Format still reads as separate dark
panels, and wide-screen negative space often becomes dead space.

Stage 4H therefore changes **composition, scale and relationships**, not visual
vocabulary.

The primary governing rule is:

> **Fill space with scale, relationships and depth — not with more components.**

The existing filters remain valid:

> **Visual complexity is allowed when it creates spatial hierarchy,
> interaction, orientation, or a memorable product object.**

> **If an element exists only to make an empty region look more interesting,
> remove it.**

Negative space is not a bug. Dead space is different: it is space that has no
useful compositional relationship to content, hierarchy or depth. Stage 4H
should convert dead space into intentional composition, not decorate it.

## What Stage 4F taught us

Stage 4F successfully removed:

- cartoon/editorial illustration as a public identity;
- giant stage numerals;
- forced asymmetry and designed imperfection;
- multicolor section fields;
- slogan-heavy copy;
- narrative scroll choreography;
- Bento/Mosaic composition;
- art-panel authentication layouts.

Those removals remain valid.

However, a completely flat content-first implementation can read like a polished
wireframe. SPCase still needs a recognizable visual identity and stronger UX
hierarchy. The solution is **not** to reintroduce decorative scenes. It is to
use restrained pseudo-3D, material depth and interactive spatial relationships
where the underlying content benefits from them.

## What Stage 4G taught us

Stage 4G proved that pseudo-3D can belong in SPCase without restoring the old
illustration grammar. Keep:

- one Z2 signature moment rather than site-wide 3D;
- graphite material planes;
- restrained red;
- fine-pointer tilt only as progressive enhancement;
- Z0 schedule/FAQ/auth contrast;
- no new visual runtime dependency.

The problem is **confidence and integration**, not the chosen language. Stage
4H must address four specific weaknesses:

1. Hero spatial identity is too small relative to the desktop viewport and can
   read as several stacked cards rather than one SPCase-specific object.
2. Format uses depth technically, but visually still resembles three separate
   dark panels.
3. The global desktop content footprint is too narrow/timid in places; empty
   viewport area does not always participate in hierarchy.
4. Information-rich sections such as Schedule and FAQ can carry more visual
   weight through typography, alignment and width without adding decoration.

## Composition and density

### Scale before count

When a surface feels empty, first test whether existing important content is too
small or too weakly related. Prefer enlarging/recomposing one meaningful system
over adding more objects.

Do not solve density by adding:

- more cards;
- more metadata than the product needs;
- decorative labels;
- abstract geometry;
- filler SVG;
- extra pseudo-3D layers;
- slogans or invented content.

### Wide layout, narrow reading

The overall desktop composition may use a wide canvas (roughly 1500–1600 px on
large screens when viewport/gutters permit), while prose and form reading
columns remain much narrower. `max-width` for the page and readable line length
are separate decisions.

Do not center a 700–1100 px visual island inside a 1920 px viewport by default.
Do not stretch body text simply because the layout is wide.

### Negative space vs dead space

Negative space is intentional when it:

- separates hierarchy;
- frames a signature object;
- creates tension between large and small information;
- allows spatial depth to read;
- makes a later dense section feel distinct.

Dead space is empty area created because important content is underscaled,
unrelated or trapped in an unnecessarily narrow container. Fix dead space by
changing scale, alignment, overlap/depth relationships or information layout.
Do not fix it with filler.

### Section density variation

Different surfaces may occupy very different amounts of space. Hero may be
large and spatial; FAQ may be wide and typographic; auth may be compact;
schedule may be dense. This variation is desirable. Do not normalize every
section to the same vertical padding or visual loudness.

## Dark material palette

The default canvas remains dark. Use a narrow neutral material range plus one
primary red accent.

A practical starting relationship:

```text
canvas                 #090D13 to #0D1118
plane / material        #0F151D
plane +1                #151C25
edge / separator        #29313C
primary text            #ECEAE4
muted text              #8993A2
deep red                #C83A32
accessible red text     same hue, lighter tone as required for contrast
```

Exact values may be calibrated during implementation and human review.

The visual system should still read as mostly dark neutral material. Do not
restore mustard/cyan/coral as large public fields.

### Gradients and shadows

Stage 4F broadly rejected gradients and shadows because they were likely to
become dark-SaaS decoration. The current spatial direction permits them **only
as material-lighting primitives**.

Allowed examples:

- a very low-contrast directional gradient that makes a spatial plane readable;
- a soft shadow cast by one meaningful layer onto another;
- a subtle edge highlight that communicates thickness or selection.

Not allowed:

- gradient page backgrounds;
- purple/blue startup gradients;
- ambient neon glow;
- luminous borders;
- glassmorphism;
- shadows on every container;
- “premium black” chrome.

If removing a gradient/shadow leaves the information hierarchy unchanged, it is
probably decorative and should be removed.

## Surface and depth model

The page itself remains the primary surface. Pseudo-3D is an enhancement, not
the default container model.

Use three depth levels:

### Z0 — interface

Normal 2D UI:

- body copy;
- navigation;
- forms;
- FAQ;
- controls;
- most product data;
- ordinary links and buttons.

Z0 must remain complete and attractive without any 3D effect.

### Z1 — structural depth

Shallow depth used to clarify a relationship or state:

- phase/step relationships;
- selected/active objects;
- schedule day layers where useful;
- a small foreground/background distinction;
- direct hover/focus response on an already meaningful object.

Typical implementation should feel shallow: small `translateZ`, small offsets,
very small rotation. Z1 must never resemble floating cards in space.

### Z2 — signature spatial object

One or at most a few high-salience public moments may use stronger pseudo-3D.
The homepage hero is the primary candidate.

A Z2 object should:

- carry truthful event/product information, or strongly reinforce event
  identity;
- be understandable when completely static;
- occupy a meaningful compositional role rather than filling empty space;
- respond subtly to input, not perform continuously;
- collapse gracefully to a simpler 2D/mobile presentation.

Do not spread Z2 treatment across every section.

## Pseudo-3D implementation constraints

Stage 4H continues to use browser-native and already-installed primitives:

- CSS `perspective`;
- `transform-style: preserve-3d`;
- layered DOM/CSS surfaces;
- SVG only for interface/brand geometry where justified;
- existing Motion values/springs for restrained interaction.

Do **not** add Three.js, React Three Fiber, WebGL frameworks, shaders or another
runtime visual dependency in Stage 4H.

True WebGL may be considered later only if a concrete interaction cannot be
expressed convincingly with the existing stack and receives explicit human
approval.

## Spatial motion

Spatial motion is allowed when it reveals hierarchy or direct manipulation.

Good examples:

- a signature hero object tilts a few degrees toward fine-pointer movement;
- an active stage plane advances slightly while neighboring planes recede;
- a schedule layer transitions forward when the user deliberately changes
  context;
- a pressed control gains/reduces perceived depth.

Avoid:

- idle floating;
- constant looping rotation;
- large cursor-following travel;
- scroll-driven depth for decoration;
- dramatic camera moves;
- section-by-section reveal choreography.

As a general target for signature pointer tilt, remain around a few degrees,
not tens of degrees. Interaction should feel physical, not toy-like.

`prefers-reduced-motion` must produce a complete static state with no loss of
information.

## Typography

Typography remains a primary identity layer, but Stage 4H does not return to
ubiquitous poster type.

- Hero display type may be large enough to establish event presence.
- Section headings should normally be much quieter than the hero.
- Factual metadata can be visually strong without becoming KPI cards.
- Dates, times and numbers may use tabular figures.
- Manrope may remain the main family unless human review later identifies a
  specific typographic problem.
- Do not add a second display face merely to manufacture character.

Large type is allowed because a hero deserves hierarchy, not because every
section needs a “wow” moment.

## Content and visual objects

Real information remains preferred over decorative filler.

A spatial object may use real metadata such as:

- event/year;
- city;
- registration deadline;
- team size;
- competition phase labels;
- schedule day/time;
- case/submission state in future product surfaces.

Do not invent fake IDs, LIVE/NOW states, team counts, scores or deadlines to
make a spatial object look richer.

A meaningful spatial object is different from an illustration. It participates
in hierarchy or interaction and may carry data. Do not simply model a gear,
flag, podium, floating document or abstract 3D blob.

## Homepage

The homepage is the main high-salience public surface. Stage 4H should make it
feel confident on a wide desktop without returning to decorative illustration.

The hero should contain:

- clear SPCase/championship identity;
- concise factual lead;
- primary action;
- truthful event/registration facts where available;
- one integrated signature spatial composition.

### Hero scale

The hero is the one place where display typography may become substantially
larger than the rest of the site. It should create hierarchy, not recreate the
Stage 4E poster system.

On large screens the signature object may occupy roughly 35–45% of the useful
composition if that produces a better balance. It should not remain a small
artifact surrounded by unrelated empty viewport.

### One spatial object, not a card stack

The signature composition should read as **one physical/system object** — a
case chassis, assembly, carrier or other integrated plane system — rather than
several independent floating cards.

Real metadata may be embedded into faces/layers of the same object. Supporting
backplanes may create depth, but every plane should contribute to one coherent
silhouette/relationship.

Do not add a second hero illustration or decorative background object to make
the scene larger.

A left-content/right-object geometry is allowed when the two sides feel like
one composition. It is not automatically rejected; the failure mode is a small
"cool object" placed beside unrelated text.

Static composition must be strong under reduced motion. Pointer tilt remains a
minor physical response, never the source of visual weight.

## Format

Explain how the championship works in direct language.

Each stage communicates:

- stage name;
- date/window when authoritative;
- what participants do;
- what outcome moves them forward.

On wide desktop, Stage 4H may represent the sequence as **one connected shallow
spatial progression**: shared chassis/steps/planes whose geometry makes the
ordering legible.

Avoid three independent bounded panels that merely differ by `translateZ` or
shadow. If the 3D transform is removed, the sequence must still read clearly.

Do not restore:

- giant stage numerals;
- poster offsets;
- flags/gears;
- large alternating color fields;
- hover spectacle on non-interactive stages.

On mobile/touch, flatten to a straightforward vertical sequence.

## Schedule

Schedule data is still the primary graphic material.

Stage 4H should increase its visual footprint primarily through:

- aligned time columns;
- wider rules/rows;
- stronger date hierarchy;
- deliberate use of horizontal space;
- larger but restrained numeric/time scale where it improves scanning.

Spatial treatment remains optional and must improve orientation. Possible uses:

- day groups treated as shallow layers;
- the selected/current day advancing when a truthful state exists;
- a controlled transition between day contexts.

Do not hide schedule information behind a carousel-like 3D interaction. The
full schedule must remain directly scannable, keyboard accessible and strong in
2D/mobile layouts.

The schedule should look fuller because real data is composed with confidence,
not because each event became a card.

## FAQ

FAQ remains deliberately quiet and predominantly Z0.

Do not add 3D merely for consistency with the hero. Expansion/collapse is a
state transition, not a spectacle surface.

On wide screens FAQ may use much more of the page width as a typographic
surface: long separators, clear question rhythm and a comfortable answer
measure. Do not constrain it to a small centered island merely to preserve
emptiness.

## Authentication

Login and registration remain predominantly Z0 functional forms.

A single restrained spatial identity element may be used only if it does not
compete with the form and can disappear on small screens without changing the
workflow.

Inputs, labels, validation, focus and submit controls must remain conventional,
stable and predictable.

No art panel, glass panel or “floating form card” composition.

## Product surfaces

Future USER/JURY product surfaces inherit the dark material language but use
spatial depth primarily for **state and hierarchy**, not decoration.

Examples that may become valid later:

- active work artifact slightly foregrounded;
- locked/finalized artifact visually settles into a non-interactive depth state;
- selected case/submission layer separated from supporting material;
- consequential transitions communicate a real state change.

Do not design these Phase 5 surfaces during Stage 4H.

ADMIN remains predominantly Z0 and utilitarian.

## Navigation

Navigation should be legible and quiet. Existing continuity motion may remain
if it is useful.

Do not make navigation itself the signature pseudo-3D object. Avoid HUD chrome,
hover spectacle and floating navigation panels.

## Anti-slop contract

Reject as dominant grammar:

- generic dark SaaS cards;
- a hero made from several visibly independent spatial cards;
- a signature object too small to establish a relationship with its viewport;
- separate dark Format panels pretending to be one progression through tiny depth deltas;
- filling dead space by adding more components instead of recomposing scale;
- cyberpunk/HUD/terminal aesthetics;
- neon or glow;
- glassmorphism;
- purple/blue gradients;
- floating abstract cubes/blobs;
- rotating 3D logos;
- Apple-Vision-style floating glass panels;
- “premium industrial” decoration with meaningless screws/grids/coordinates;
- pseudo-3D applied to every section;
- illustration disguised as 3D;
- constant parallax or cursor-following scenes;
- Swiss poster tropes from Stage 4E;
- slogan copy replacing real information.

Spatial design must not become a new template grammar.

## Responsive behavior

Mobile is not expected to reproduce desktop perspective literally.

At 375 px and 320 px:

- flatten or reduce depth when it improves clarity;
- preserve the same information hierarchy;
- avoid horizontal overflow from transformed elements;
- keep controls at least 44 px;
- never require hover/fine pointer;
- ensure spatial surfaces do not leave large dead gaps when hidden/flattened;
- prioritize content over preserving desktop spectacle.

## Accessibility

- Spatial state must never be communicated only by depth, color or motion.
- Keyboard/focus behavior must remain explicit.
- Text contrast must meet the project accessibility baseline.
- Reduced-motion users receive a stable equivalent composition.
- Pointer-driven effects are progressive enhancement only.
- Transform/perspective must not make text difficult to read; important body
  copy should generally remain on readable planes.

## Decision test

Before adding a visual element, ask in order:

1. Does it carry real content, state or interaction?
2. Does depth make the relationship easier to understand or remember?
3. Is the static/reduced-motion/mobile version still complete?
4. Could the same purpose be served more clearly with typography/spacing?
5. Is this starting to look like a recognizable 3D-design trope?

If answers 1–3 are weak, do not add it.
