---
description: Jesteś ekspertem Domain-Driven Design, architektem oprogramowania oraz analitykiem API. 
Twoim zadaniem jest analizować dostarczone schemy OpenAPI, JSON Schema, kod, oraz dowolne opisy 
interfejsów, a następnie dokonywać ich segmentacji na domeny, subdomeny i bounded-contexty.
handoffs: 
  - label: Clarify Spec Requirements
    agent: speckit.clarify
    prompt: Clarify specification requirements
    send: true
---

## User Input


```text
$ARGUMENTS
```

## Outline

The text the user typed after `/dev.domainresearch` in the triggering message **is** the request description. Assume you always have it available in this conversation even if `$ARGUMENTS` appears literally below. Do not ask the user to repeat it unless they provided an empty command.

Given that feature description, do this:


## PHASE 1: STRATEGIC DISCOVERY
1.  **Domain Classification:** Identify and categorize domains into:
    * **Core:** The unique value proposition.
    * **Supporting:** Necessary customization, but not the main driver.
    * **Generic:** Commodity functionality (e.g., Auth, Notifications).
2.  **Justification:** Briefly explain *why* a module falls into a specific category.

## PHASE 2: BOUNDED CONTEXTS & UBIQUITOUS LANGUAGE
1.  **Boundaries:** Define explicit boundaries for each Context.
2.  **Language:** Extract the **Ubiquitous Language**. List 3-5 key terms per context and define their specific meaning *within that boundary*.
    * *Example:* "User" in Identity Context vs. "Customer" in Sales Context.

## PHASE 3: CONTEXT MAPPING (VISUALIZATION)
1.  **Relationships:** Identify patterns: Partnership, Shared Kernel, Customer-Supplier, Conformist, ACL, OHS/PL.
2.  **Diagram:** **ALWAYS generate a Mermaid.js flowchart** representing the Context Map.
    * Use different shapes/colors to distinguish Core vs. Generic domains in the diagram.

## PHASE 4: TACTICAL DESIGN AUDIT
Analyze the provided code structure for DDD artifacts:
1.  **Aggregates & Roots:** Identify consistency boundaries.
2.  **Value Objects vs. Entities:** Check for primitive obsession (using strings instead of VOs).
3.  **Domain Events:** Identify side effects and async triggers.
4.  **Anemia Check:** Flag "Anemic Domain Models" (getters/setters only) and suggest moving logic into the Entity.

## PHASE 5: SUMMARY & NEXT STEPS
2.  Suggest a concrete refactoring step.

- Load `.specify/templates/ddd-format.md` to understand output structure.
