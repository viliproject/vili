import _ from 'underscore';

export default function(template, variables) {
    variables = variables || {};
    var seenVariables = {},
        usedVariables = [],
        missingVariables = [];

    var re = /\{([a-zA-Z0-9_:]+)\}/g;
    var match;
    while((match = re.exec(template)) !== null) {
        var key = match[1];
        var defaultValue;
        var splitKey = key.split(':');
        if (splitKey.length > 1) {
            key = splitKey[0];
            defaultValue = splitKey[1];
        }
        var value = variables[key] || defaultValue;
        if (value !== undefined) {
            if (!seenVariables[key]) {
                usedVariables.push({key: match[1], value: value});
            }
            seenVariables[key] = value;
        } else {
            missingVariables.push(match[1]);
        }
    }
    var populated = template.replace(/\{[a-zA-Z0-9_:]+\}/g, function(match) {
        var defaultValue = '';
        var varName = match.substring(1, match.length - 1);
        var splitVar = varName.split(':');
        if (splitVar.length > 1) {
            varName = splitVar[0];
            defaultValue = splitVar[1];
        }
        var value = variables[varName];
        if (value !== undefined) {
            return value;
        } else {
            return defaultValue;
        }
    });
    return {
        template: template,
        usedVariables: usedVariables,
        missingVariables: missingVariables,
        populated: populated,
        valid: _.isEmpty(missingVariables),
    };
}
