import _ from 'underscore';

export default function(template, variables) {
    variables = variables || {};
    var seenVariables = {},
        usedVariables = [],
        missingVariables = [];

    var re = /\{([a-zA-Z0-9_]+)\}/g;
    var match;
    while((match = re.exec(template)) !== null) {
        var key = match[1],
            value = variables[key];
        if (value !== undefined) {
            if (!seenVariables[key]) {
                usedVariables.push({key: match[1], value: value});
            }
            seenVariables[key] = value;
        } else {
            missingVariables.push(match[1]);
        }
    }
    var populated = template.replace(/\{[a-zA-Z0-9_]+\}/g, function(match) {
        var value = variables[match.substring(1, match.length - 1)];
        if (value !== undefined) {
            return value;
        } else {
            return '';
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
