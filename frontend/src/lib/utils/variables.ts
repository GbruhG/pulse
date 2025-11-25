// Variable substitution utilities
// Handles replacing {{variable}} patterns in text with actual values

export function substituteVariables(
    text: string,
    variables: Record<string, string>
): string {
    if (!text) return text;

    // Replace {{variable_name}} with actual values
    return text.replace(/\{\{([^}]+)\}\}/g, (match, varName) => {
        const trimmedName = varName.trim();
        return variables[trimmedName] !== undefined ? variables[trimmedName] : match;
    });
}

export function substituteInObject(
    obj: any,
    variables: Record<string, string>
): any {
    if (typeof obj === 'string') {
        return substituteVariables(obj, variables);
    }

    if (Array.isArray(obj)) {
        return obj.map(item => substituteInObject(item, variables));
    }

    if (obj && typeof obj === 'object') {
        const result: any = {};
        for (const [key, value] of Object.entries(obj)) {
            result[key] = substituteInObject(value, variables);
        }
        return result;
    }

    return obj;
}

export function extractVariables(text: string): string[] {
    const matches = text.matchAll(/\{\{([^}]+)\}\}/g);
    return Array.from(matches, m => m[1].trim());
}