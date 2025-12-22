/**
 * Deep comparison utility to find changed paths between two objects
 */
export function findChangedPaths(oldObj: any, newObj: any, currentPath: string = ""): Set<string> {
  const changedPaths = new Set<string>();

  // Handle null/undefined cases
  if (oldObj === newObj) return changedPaths;
  if (oldObj == null || newObj == null) {
    if (oldObj !== newObj && currentPath) {
      changedPaths.add(currentPath);
    }
    return changedPaths;
  }

  // Handle primitive types
  const oldType = typeof oldObj;
  const newType = typeof newObj;

  if (oldType !== newType) {
    if (currentPath) changedPaths.add(currentPath);
    return changedPaths;
  }

  if (oldType !== "object") {
    if (oldObj !== newObj && currentPath) {
      changedPaths.add(currentPath);
    }
    return changedPaths;
  }

  // Handle arrays
  const oldIsArray = Array.isArray(oldObj);
  const newIsArray = Array.isArray(newObj);

  if (oldIsArray !== newIsArray) {
    if (currentPath) changedPaths.add(currentPath);
    return changedPaths;
  }

  if (oldIsArray) {
    // Check array length changes
    if (oldObj.length !== newObj.length && currentPath) {
      changedPaths.add(currentPath);
    }

    // Check each array element
    const maxLength = Math.max(oldObj.length, newObj.length);
    for (let i = 0; i < maxLength; i++) {
      const elementPath = currentPath ? `${currentPath}.${i}` : String(i);
      if (i >= oldObj.length || i >= newObj.length) {
        changedPaths.add(elementPath);
      } else {
        const elementChanges = findChangedPaths(oldObj[i], newObj[i], elementPath);
        elementChanges.forEach((path) => changedPaths.add(path));
      }
    }
    return changedPaths;
  }

  // Handle objects
  const allKeys = new Set([...Object.keys(oldObj), ...Object.keys(newObj)]);

  for (const key of allKeys) {
    const keyPath = currentPath ? `${currentPath}.${key}` : key;

    if (!(key in oldObj)) {
      changedPaths.add(keyPath);
    } else if (!(key in newObj)) {
      changedPaths.add(keyPath);
    } else {
      const propertyChanges = findChangedPaths(oldObj[key], newObj[key], keyPath);
      propertyChanges.forEach((path) => changedPaths.add(path));
    }
  }

  return changedPaths;
}

/**
 * Deep clone an object
 */
export function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== "object") {
    return obj;
  }

  if (obj instanceof Date) {
    return new Date(obj.getTime()) as any;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => deepClone(item)) as any;
  }

  const clonedObj: any = {};
  for (const key in obj) {
    if (Object.prototype.hasOwnProperty.call(obj, key)) {
      clonedObj[key] = deepClone(obj[key]);
    }
  }

  return clonedObj;
}
