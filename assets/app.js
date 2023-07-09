import { ref } from 'vue'
import Cookies from "js-cookie"
export default {
  setup() {
    const fullname = Cookies.get("fullname") || ""
    const email = Cookies.get("email")
    const avatar = Cookies.get("avatar")
    const userid = Cookies.get("userid")
    console.log("fullname", fullname, fullname === "")
    if (fullname === "") {
      location.href = `/web/login?next=${btoa(location.href)}`
    }
    const logout = ()=> {
      location.href = `/web/logout?next=${location.href}`
    }
    return {
      fullname,
      email,
      avatar,
      userid,
      logout
    }
  },
}
