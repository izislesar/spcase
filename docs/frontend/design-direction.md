# Frontend design direction

> **Status: approved stable visual constitution for Stage 4G and subsequent
> product-surface work unless superseded by human review.**
> Current human verdicts live in `visual-acceptance.md`; product/interaction
> semantics live in `experience-model.md`.

## North star

SPCase is a **dark spatial competition interface**.

It should feel like a real championship with a strong digital presence, not a
SaaS dashboard, a creative-agency portfolio, a Swiss-poster exercise or a
3D-showcase website.

Stage 4F established the correct neutral foundation: dark canvas, factual copy,
low decorative noise and much less motion. Human review accepted that cleanup
as a foundation but found the rendered result **visually incomplete**. Stage 4G
adds depth, spatial hierarchy and a small number of memorable visual moments
without restoring illustration-led design.

The governing rule is:

> **Visual complexity is allowed when it creates spatial hierarchy,
> interaction, orientation, or a memorable product object.**

The older Stage 4F test remains useful:

> **If an element exists only to make an empty region look more interesting,
> remove it.**

These rules are complementary. Stage 4G is not a return to decoration; it is a
controlled increase in visual richness.

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
become dark-SaaS decoration. Stage 4G permits them **only as material-lighting
primitives**.

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

Stage 4G should first use browser-native and already-installed primitives:

- CSS `perspective`;
- `transform-style: preserve-3d`;
- layered DOM/CSS surfaces;
- SVG only for interface/brand geometry where justified;
- existing Motion values/springs for restrained interaction.

Do **not** add Three.js, React Three Fiber, WebGL frameworks, shaders or another
runtime visual dependency in Stage 4G.

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

Typography remains a primary identity layer, but Stage 4G does not return to
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

The homepage should combine the factual clarity of Stage 4F with one strong
spatial identity moment.

The hero should contain:

- SPCase/championship identity;
- clear primary title;
- truthful location/date/registration information that the application already
  has authority to show;
- one obvious primary action;
- restrained secondary navigation/context;
- one signature spatial composition when it materially improves the page.

The hero must not regress to `headline left + decorative object right` by
default. If a spatial object occupies the right side, it must carry information
or participate in the page hierarchy rather than function as illustration.

A good static screenshot must still work if pointer interaction is disabled.

## Format

`Формат чемпионата` should remain information-first, but it no longer needs to
be visually flat.

Allowed:

- a shallow spatial sequence that clarifies progression;
- layered/stepped planes tied to the actual three stages;
- direct interaction that brings the focused stage forward;
- restrained depth cues.

Not allowed:

- returning to giant `01/02/03` as decoration;
- gear/flag scenes;
- colored poster bands;
- deliberately broken grid rules;
- three equal SaaS cards;
- a complex 3D scene whose only value is spectacle.

On mobile, the information should reduce naturally to a clear vertical sequence
without depending on perspective.

## Schedule

Schedule data is still the primary graphic material.

Spatial treatment is optional and must improve orientation. Possible uses:

- day groups treated as shallow layers;
- the selected/current day advancing when a truthful state exists;
- a controlled transition between day contexts.

Do not hide schedule information behind a carousel-like 3D interaction. The
full schedule must remain directly scannable, keyboard accessible and strong in
2D/mobile layouts.

## FAQ

FAQ remains deliberately quiet and predominantly Z0.

Do not add 3D merely for consistency with the hero. Expansion/collapse is a
state transition, not a spectacle surface.

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

Do not design these Phase 5 surfaces during Stage 4G.

ADMIN remains predominantly Z0 and utilitarian.

## Navigation

Navigation should be legible and quiet. Existing continuity motion may remain
if it is useful.

Do not make navigation itself the signature pseudo-3D object. Avoid HUD chrome,
hover spectacle and floating navigation panels.

## Anti-slop contract

Reject as dominant grammar:

- generic dark SaaS cards;
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
