# Frontend design direction

> **Status: approved stable visual constitution for the next frontend iteration.**
> This document defines enduring visual rules, not a snapshot of a particular
> implementation. Current human verdicts live in `visual-acceptance.md`;
> product/interaction semantics live in `experience-model.md`.

## North star

SPCase is a **dark, content-first competition interface**.

It should feel like a real championship that happens to have a digital
interface, not like a designer's concept for a championship website.

The governing sentence is:

> **Nothing exists solely to make the page look designed.**

A second test applies to every decorative or compositional decision:

> **If an element cannot justify its existence without the phrase “visual
> interest”, remove it.**

The target is quiet confidence, not visual spectacle. Character should emerge
from real content, typography, information hierarchy, state, spacing and the
natural differences between surfaces.

## What Stage 4E taught us

The previous direction successfully moved away from generic SaaS, but it
created another recognizable AI/design-template grammar: oversized stage
numbers, authored grid breaks, flat symbolic illustrations, large primary
color blocks, pseudo-editorial rules and deliberate asymmetry.

Those devices are no longer approved as identity by default.

Do not try to fix that result by inventing more unusual visual tricks. Stage
4F follows a **delete before designing** rule: remove anything whose main job
is aesthetic signaling, then rebuild only what the content and interaction
actually require.

## Dark canvas

The default public/product canvas is dark rather than cream/off-white.

Use a narrow palette. A good starting set is:

```text
canvas             #0D1118
surface             #131923
raised/selected     #1A2230
primary text        #E8E6DF
muted text          #8E98A8
rules               rgba(232, 230, 223, 0.16)
primary accent      #C83A32
optional cool tone  #60758A
```

Exact values may be calibrated during implementation and human review, but the
relationship is authoritative:

- roughly 80–90% dark canvas;
- most remaining visual information is warm light text and subtle rules;
- one primary accent, currently deep red;
- any secondary hue is muted and rare.

Do not use mustard, cyan/turquoise and coral as simultaneous large identity
fields. Existing tokens may remain during migration, but the public visual
system must not depend on a four-color primary palette.

Dark does **not** mean:

- cyberpunk;
- developer-tool aesthetic;
- terminal cosplay;
- luxury-black marketing;
- neon;
- glowing edges;
- glass panels;
- blue/purple gradients;
- dot-grid backgrounds;
- fake technical HUD labels.

## Surface model

The page itself is the primary surface.

Do not solve dark mode by stacking slightly lighter cards on a dark background.
Use containers only when the content is genuinely a self-contained object or
requires a bounded interactive region.

Prefer separation through:

- whitespace;
- typography;
- alignment;
- thin rules;
- density changes;
- occasionally a subtle surface shift;
- rarely the red accent.

A section does not need a background change simply because the previous
section ended.

## Typography

Typography carries most of the identity.

- Manrope remains available and may remain the primary face; it is not a
  license for ubiquitous oversized bold display text.
- Do not make every section headline a poster headline.
- Prefer a wider range of useful scales: quiet labels, readable body copy,
  medium editorial headings and occasional large display type.
- Large text must be justified by information hierarchy, not by a desire to
  make a section feel designed.
- Real data such as dates, time, scores and identifiers may use tabular figures
  or a restrained system-mono role when useful.
- Do not add another display font merely to manufacture character.

Avoid slogan-like headings where a direct label is clearer. For example,
prefer `Формат чемпионата` over copy such as `Три этапа. Одна сильная работа.`
when the latter adds no information.

## Composition

Do not design imperfection.

The previous `90% discipline / 10% disobedience` rule is retired because it
encouraged agents to manufacture a checklist of grid escapes, crops and broken
rules.

Instead:

- keep a coherent underlying layout;
- allow different content to produce naturally different density and spacing;
- do not normalize every section to the same padding, height or component
  anatomy;
- do not force symmetry;
- do not force asymmetry;
- do not add a grid violation just to demonstrate authorship;
- leave empty space empty when it has no content role.

Character should be difficult to reduce to a list of "design tricks".

## Content is the graphic

Prefer real information over decorative filler.

Examples of useful graphic material:

- city and venue;
- event dates;
- registration deadline;
- team size;
- current lifecycle state when authoritative;
- schedule times;
- case title;
- submission deadline;
- score/ranking;
- file metadata;
- participant roles.

Do not replace empty space with a gear, flag, sheet stack, podium, machine,
speech bubble or abstract symbol merely because the layout feels sparse.

A truthful `02–04.10.26` can be more visually valuable than an illustration.

## Illustration and decorative graphics

Illustration is no longer a default part of the public identity.

For Stage 4F:

- remove the current cartoon/editorial scene language from dominant public
  composition;
- do not add replacement illustration systems;
- do not create semantic gears/flags/stamps/podiums merely to preserve the old
  concept;
- use icons only where an interface convention or action genuinely benefits
  from one;
- use the SPCase mark/logo if and when a real brand asset exists;
- prefer text, data and structure over decorative SVG.

Future illustration can be reintroduced only after a specific human-approved
need, not as a default solution to empty space.

## Color semantics

Color must earn its use.

Deep red is the primary accent. Good candidates include:

- primary action;
- current/high-priority state;
- deadline emphasis;
- a consequential result or lifecycle moment.

Do not make an entire section red merely to create visual rhythm. Do not use a
second or third bright color simply to differentiate neighboring sections.
State must never rely on color alone.

## Public surfaces

Public pages should be calm, direct and event-specific.

The homepage should prioritize truthful competition content such as:

- SPCase identity;
- city/location;
- dates;
- registration deadline/state;
- team format;
- primary action;
- schedule/FAQ only when they are useful next information.

A strong public page may contain large areas with only typography and real
information.

Do not default to:

- hero art panels;
- hero illustration;
- editorial collage;
- equal feature blocks;
- giant stage numerals;
- color-field section rhythm;
- agency-style headline slogans;
- decorative section labels such as `01 · ФОРМАТ` when they add no navigation
  or information value.

## Format/stages

The championship format is content, not a visual showcase.

Present the stages plainly and clearly. Small numbers are acceptable when they
help sequence the stages, but numbers must behave as metadata rather than
dominant decoration.

Prefer actual stage names, dates and explanatory text over a custom visual
scene for every stage.

## Schedule

Schedule is one of the places where information itself can produce a strong
visual composition.

Use:

- date grouping;
- aligned times;
- event titles;
- locations/details;
- current/next state when authoritative;
- rules and spacing where they improve scanning.

Do not decorate the schedule with unrelated artwork. Motion is only useful
when it communicates temporal state or interaction.

## Auth surfaces

Login and registration are functional surfaces.

They should be especially restrained:

- one dark canvas;
- clear title/context;
- stable form fields;
- direct primary action;
- concise useful registration/deadline information where truthful.

No art panel is required. Do not add illustration, parallax or spectacle to
make auth feel "on brand".

## Product surfaces

USER and JURY surfaces should derive visual strength from operational
information rather than marketing composition.

Prefer:

- plain hierarchy;
- lists/tables where appropriate;
- explicit deadlines and state;
- files and metadata;
- clear next action;
- thin rules and restrained surface changes;
- minimal decoration;
- minimal motion.

Do not turn product screens into Bloomberg/terminal cosplay. Monospace,
technical IDs and compact metadata are tools, not identity.

## Admin surfaces

ADMIN is primarily utilitarian. Prioritize correctness, scanability and safe
operation. Branding should be quiet.

## Motion

Motion is now deliberately sparse.

Preferred reasons:

1. navigation continuity;
2. state transition;
3. direct interaction feedback;
4. truthful temporal progression where static representation is insufficient.

Public scroll reveal, decorative parallax, floating art, scroll drift and
motion added purely to make a page feel premium are not approved defaults.

A static screenshot should already contain the intended hierarchy and
character.

Every motion surface must have a complete reduced-motion/static path.

## Anti-AI / anti-template contract

Reject the design when its identity is mainly one or more of these patterns:

- Swiss/editorial agency-site imitation;
- giant `01 / 02 / 03` numerals as decoration;
- label patterns such as `01 · SECTION` used for style rather than utility;
- deliberately broken grid rules;
- designed whitespace imbalance as a signature trick;
- gear/flag/document/podium illustration motifs;
- large primary-color rectangles used as section identity;
- Bento/equal-card composition;
- KPI-card dashboards;
- universal rounded dark cards;
- glass/glow/gradient/neon dark UI;
- fake terminal/HUD aesthetics;
- generic "premium" reveal choreography;
- slogan copy generated to make ordinary information sound profound;
- decorative marks, stamps, registration symbols or pseudo-print artifacts
  without a real product purpose.

A primitive is not banned because AI systems sometimes use it. The question is
whether the primitive is justified by the actual content or interaction.

## The deletion test

Before adding or retaining any nonessential visual element, answer:

1. What information does it communicate?
2. What interaction does it support?
3. What hierarchy does it clarify?
4. Would the page become less understandable or less identifiable without it?

If the only defensible answer is "it adds visual interest", remove it.

## Mobile

Mobile is a first-class composition, but it should become simpler rather than
more theatrical.

Core functionality must work:

- from 320 px upward;
- on touch-only devices;
- without hover;
- with reduced motion;
- with touch targets at least 44 px where interactive;
- without horizontal overflow caused by decorative layout tricks.

When desktop whitespace or alignment does not translate naturally to a small
screen, recompute it for clarity instead of preserving an art-directed gesture.
