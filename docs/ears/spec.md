# Spec

## 1.0 Formal Morphological Constraints for Software Systems

Within the domain of software architecture, unconstrained natural language introduces severe systemic vulnerabilities, including logic omissions, algorithmic ambiguity, and untestability. The Easy Approach to Requirements Syntax (EARS) rectifies this by imposing a strict, rule-based temporal logic on textual requirements, ensuring that software behaviors are deterministic and precisely mapped to system execution states.

The architectural integrity of a software requirement in EARS relies on an immutable grammatical sequence: `<optional preconditions> <optional trigger> the <system name> shall <system response>`.

For software development, the `<system name>` directly references the computational actor (e.g., the specific microservice, class, or operating system), while the `<system response>` dictates the mandatory algorithmic execution or output. Furthermore, complex boolean logic within software requirements must be carefully partitioned; terms like "and/or" must be explicitly separated into exclusive conditions to avoid combinatorial explosion and ensure precise test coverage.

## 2.0 Software-Specific EARS Pattern Specification

By constraining vocabulary into five distinct patterns, EARS categorizes software behaviors into predictable architectural constructs. The following specification restricts all instantiations strictly to software engineering contexts.

### 2.1 Ubiquitous Software Requirements

- **Architectural Intent:** Defines fundamental, global software properties, baseline constraints, or continuous background services that do not require an external stimulus to execute.
- **Formal Syntax:** `The <system name> shall <system response>.`
- **Software Examples:**
  - *Language/Platform Constraint:* `The software shall be written in Java.`
  - *Deployment Property:* `The software package shall include an installer.`
  - *Continuous Access Logic:* `The system shall allow the admin to verify new brands.`

### 2.2 Event-Driven Software Requirements

- **Architectural Intent:** Governs discrete, strictly defined API invocations, asynchronous message receptions, or user interface interactions. A specific system response is mandated *when and only when* the triggering event is detected.
- **Formal Syntax:** `When <trigger>, the <system name> shall <system response>.`
- **Software Examples:**
  - *Enterprise Transaction:* `When an Order is shipped and Order Terms are not 'Prepaid', the system shall create an Invoice.`
  - *User Interface Event:* `When the user selects the caller count from the menu, the software shall display a count of the number of participants in the audio call in the UI.`
  - *Asynchronous Workflow:* `When the driver has accepted the package request, the system shall send the package request to the selected driver, and the rider will be notified.`

### 2.3 State-Driven Software Requirements

- **Architectural Intent:** Dictates modal software behaviors, session states, or specific runtime environments. The software response remains actively enforced exclusively for the duration that the operational state evaluates to true.
- **Formal Syntax:** `While <system state>, the <system name> shall <system response>.`
- **Software Examples:**
  - *Process State:* `While the mute button is depressed, the software shall mute the microphone.`
  - *Initialization Mode:* `While in Manufacturing Mode, the software shall boot without user intervention.`

### 2.4 Optional Feature Software Requirements

- **Architectural Intent:** Critical for Software Product Line Engineering (SPLE), this pattern manages modular configurations and feature flags, decoupling variable logic from the core application codebase.
- **Formal Syntax:** `Where <feature is included>, the <system name> shall <system response>.`
- **Software Examples:**
  - *Application Module Variability:* `Where a thesaurus is part of the software package, the installer shall prompt the user before installing the thesaurus.`
  - *E-Commerce Configuration:* `Where cash on delivery are available, the system shall allow the customer to select the payment method.`
  - *Digital Rights Management (DRM):* `Where the book is available in digital format, the software shall allow the user to download the book without charge for a trial period of 3 days.`

### 2.5 Unwanted Behaviour Software Requirements

- **Architectural Intent:** Establishes the foundational logic for exception handling, cybersecurity fault tolerance, and data validation. It articulates the precise computational mitigation required when the software encounters malformed inputs or backend failures.
- **Formal Syntax:** `If <trigger>, then the <system name> shall <system response>.`
- **Software Examples:**
  - *Data Validation Exception:* `If an invalid credit card number is entered, then the website shall display 'please re-enter credit card details'.`
  - *Integrity Check Failure:* `If the memory checksum is invalid, then the software shall display an error message.`
  - *Service Dependency Failure:* `If the alarm software detects that a sensor has malfunctioned, then the alarm software shall phone the Alarm Company to report the malfunction.`

### 2.6 Complex Software Requirements

- **Architectural Intent:** Synthesizes dense Boolean logic, fusing specific runtime states with triggering events or exceptions.
- **Formal Syntax:** `<Multiple Conditions>, the <system name> shall <system response>.`
- **Software Examples:**
  - *State and Unwanted Behavior Synthesis:* `While on DC power, if the software detects an error, then the software shall cache the error message instead of writing the error message to disk.`
  - *Security Threshold Logic:* `When more than 3 incorrect login attempts occur for a single user ID within a 30 minute period, the software shall lock the account associated with that user ID.`

## 3.0 Advanced EARS (Adv-EARS) for Automated Software Verification

To scale requirements engineering within advanced software factories, textual specifications must be parsable by automated toolchains to construct Abstract Syntax Trees (ASTs) and generate Unified Modeling Language (UML) diagrams. The Adv-EARS framework provides this strict, object-oriented translation layer.

- **Lexical Mapping:** Adv-EARS formally maps grammatical elements directly to Object-Oriented paradigms. The standard `<system name>` is refined into an `<entity>`, which explicitly corresponds to the actors or software classes interacting with the system. The `<system response>` is refined into `<functionality>`, which maps directly to executable software use cases.
- **Context-Free Grammar (CFG) Parsing:** By formatting software requirements into this highly constrained syntax, requirements documentation acts as executable data. A CFG can programmatically parse these requirements to identify system entities and automatically synthesize use case models, radically lowering the risk of human translation errors moving from requirements specification into backend architecture.

## 4.0 Software Requirements Metadata Integration

Text-based EARS statements are mathematically sound, but in enterprise software deployments, they must be embedded within a structured planning language (such as "Planguage") to ensure full lifecycle traceability, impact analysis, and verification. A rigorous software requirement record must encompass:

1. **Name:** A discrete, programmatic identifier (e.g., `Create_Invoice`).
2. **Requirement:** The strictly formatted EARS textual syntax.
3. **Rationale:** The business or architectural logic dictating the function, serving as justification for the implementation effort.
4. **Priority & Criticality:** Boolean or hierarchical metrics dictating implementation sequencing during Agile sprints or resource constraints.
5. **Traceability:** Bidirectional linkages connecting the requirement directly to its source logic, peer dependencies, and downstream software tests, ensuring seamless Change Impact Analysis (CIA).
