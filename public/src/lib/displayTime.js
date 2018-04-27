var moment = require("moment")

export default function(date, lowercase) {
  var m = date ? moment(date) : moment()
  if (lowercase) {
    return m.calendar(null, {
      lastDay: "[yesterday at] LT",
      sameDay: "[today at] LT",
      nextDay: "[tomorrow at] LT",
      lastWeek: "[last] ddd [at] LT",
      nextWeek: "ddd [at] LT",
      sameElse: "M/D/YY [at] LT",
    })
  } else {
    return m.calendar(null, {
      lastDay: "[Yesterday at] LT",
      sameDay: "[Today at] LT",
      nextDay: "[Tomorrow at] LT",
      lastWeek: "[Last] ddd [at] LT",
      nextWeek: "ddd [at] LT",
      sameElse: "M/D/YY [at] LT",
    })
  }
}
