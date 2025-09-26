# Declarative UI Patterns: Z-Index-Free Development

This document outlines the new declarative UI patterns implemented to eliminate z-index dependencies and create more maintainable, predictable interfaces.

## Overview

The traditional approach of using z-index values creates several problems:

- **Z-index wars**: Escalating values as developers try to override each other
- **Context confusion**: Elements getting stuck in the wrong stacking context
- **Hard to maintain**: Scattered z-index values throughout the codebase
- **Unpredictable behavior**: Complex interactions between positioned elements

Our new approach uses **DOM ordering** and **CSS containment** for predictable, declarative UI layering.

## Core Principles

### 1. DOM Order Determines Visual Order

Instead of fighting the natural document flow, we embrace it:

```html
<!-- Later elements appear on top -->
<div>Background content</div>
<div>UI overlays</div>
<div>Modals</div>
<div>System notifications</div>
```

### 2. Isolation Instead of Z-Index

Use `isolation: isolate` to create stacking contexts without z-index:

```css
.hover-element:hover {
  transform: scale(1.1);
  isolation: isolate; /* Creates stacking context naturally */
  /* No z-index needed! */
}
```

### 3. Portal-Based Modal System

Modals render in separate DOM trees, eliminating layering conflicts:

```tsx
// Modals render in dedicated containers
<div id="modal-primary">   <!-- Primary modals -->
<div id="modal-secondary"> <!-- Confirmation dialogs -->
<div id="modal-system">    <!-- System notifications -->
```

## Implementation Patterns

### Modal Management

#### Old Approach (Z-Index Based)

```css
.modal-overlay {
  z-index: 3000; /* Arbitrary, conflicts with other modals */
}
.confirmation-modal {
  z-index: 4000; /* Higher value needed */
}
```

#### New Approach (DOM Based)

```tsx
import { useModal } from "../components/ui/portal/ModalProvider";

const MyComponent = () => {
  const { openModal } = useModal();

  const showConfirmation = () => {
    openModal("confirm-delete", ConfirmationModal, {
      message: "Are you sure?",
      level: "secondary", // Automatically appears above primary modals
    });
  };
};
```

### Hover Effects

#### Old Approach

```css
.card:hover {
  z-index: 10; /* Can conflict with other elements */
}
```

#### New Approach

```css
.card:hover {
  isolation: isolate; /* Creates natural stacking context */
  transform: translateY(-2px); /* Visual elevation */
}
```

### UI Layout Structure

#### DOM Hierarchy for Natural Stacking

```tsx
<GameInterface>
  {/* Base layer - game board */}
  <GameBoard />

  {/* UI layer - controls and panels */}
  <BottomResourceBar />
  <LeftSidebar />
  <RightSidebar />

  {/* Overlay layer - temporary UI */}
  <CardFanOverlay />

  {/* Modal layer - managed by ModalProvider */}
  <ModalProvider>{children}</ModalProvider>
</GameInterface>
```

## Component Examples

### Creating a Z-Index-Free Modal

```tsx
import { useModal } from "../../../hooks/useModal";

const ActionButton: React.FC = () => {
  const { openModal } = useModal();

  const showActionsModal = () => {
    openModal("player-actions", ActionsModal, {
      level: "primary", // Primary modal level
    });
  };

  return <button onClick={showActionsModal}>Show Actions</button>;
};
```

### Modal Component Structure

```tsx
const MyModal: React.FC<{ onClose: () => void }> = ({ onClose }) => {
  return (
    <div className="modal-content">
      <button onClick={onClose}>Close</button>
      {/* Modal content */}

      <style jsx>{`
        .modal-content {
          /* No z-index needed - ModalProvider handles layering */
          isolation: isolate; /* Creates clean stacking context */
          background: white;
          border-radius: 8px;
          padding: 20px;
        }
      `}</style>
    </div>
  );
};
```

## Migration Guide

### Step 1: Identify Z-Index Usage

```bash
# Find all z-index usage in your codebase
grep -r "z-index" src/
```

### Step 2: Categorize Usage

- **Modals/Overlays**: Move to ModalProvider system
- **Hover Effects**: Replace with `isolation: isolate`
- **UI Layout**: Restructure DOM order
- **Background Elements**: Remove z-index, rely on DOM order

### Step 3: Replace Patterns

#### Modals

```tsx
// Before
const [showModal, setShowModal] = useState(false);
return (
  <>
    {showModal && (
      <div className="modal-overlay" style={{ zIndex: 3000 }}>
        <Modal onClose={() => setShowModal(false)} />
      </div>
    )}
  </>
);

// After
const { openModal } = useModal();
const showMyModal = () => {
  openModal("my-modal", Modal, { level: "primary" });
};
```

#### Hover Effects

```css
/* Before */
.element:hover {
  z-index: 10;
}

/* After */
.element:hover {
  isolation: isolate;
}
```

## Benefits

### 1. Predictable Behavior

- No more z-index wars
- Clear visual hierarchy through DOM structure
- Consistent stacking across components

### 2. Better Performance

- Reduced CSS complexity
- Fewer stacking context calculations
- Cleaner render tree

### 3. Improved Maintainability

- Centralized modal management
- Clear component boundaries
- Easier debugging

### 4. Accessibility Improvements

- Proper focus management through ModalProvider
- Keyboard navigation works naturally
- Screen reader compatibility

## Best Practices

### DO:

- ✅ Use DOM order for visual hierarchy
- ✅ Use `isolation: isolate` for stacking contexts
- ✅ Centralize modal management with ModalProvider
- ✅ Design components to be self-contained
- ✅ Test with keyboard navigation

### DON'T:

- ❌ Add z-index values to new components
- ❌ Fight natural document flow
- ❌ Create deeply nested stacking contexts
- ❌ Mix old and new patterns in same component
- ❌ Forget to test modal interactions

## Troubleshooting

### Modal Not Appearing Above Other Elements

- Check that ModalProvider is at the correct level in your component tree
- Ensure modal containers exist in the DOM
- Verify the correct modal level is being used

### Hover Effects Not Working

- Ensure element has `position: relative` or is otherwise positioned
- Check that `isolation: isolate` is applied correctly
- Verify there are no competing stacking contexts

### Layout Issues After Migration

- Check DOM order matches intended visual hierarchy
- Remove any remaining z-index values
- Ensure components use isolation instead of z-index

## Related Files

- `components/ui/portal/ModalProvider.tsx` - Main modal management system
- `components/ui/portal/ModalPortal.tsx` - Individual portal implementation
- `hooks/useModalStack.ts` - Modal state management hook
- `components/ui/modals/DeclarativeModalPopup.tsx` - Example implementation

## Future Improvements

1. **Automatic Z-Index Detection**: Add ESLint rules to prevent z-index usage
2. **Animation System**: Implement declarative animations for state transitions
3. **Layout Constraints**: Add system for preventing layout conflicts
4. **Developer Tools**: Create debugging tools for visual hierarchy
