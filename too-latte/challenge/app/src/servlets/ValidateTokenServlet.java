package org.bsidessf.ctf.toolatte;

import java.io.IOException;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

public class ValidateTokenServlet extends HttpServlet {
  public void doPost(HttpServletRequest httpServletRequest, HttpServletResponse httpServletResponse) throws ServletException, IOException {
    // Response response = null;
    try {
      TokenResponse response = TokenAPI.getValidateResponse(httpServletRequest.getParameter("token"));

      httpServletResponse.setStatus(HttpServletResponse.SC_OK);
      httpServletResponse.getWriter().write(response.getResponse());
      httpServletResponse.getWriter().flush();
    } catch (Exception e) {
        httpServletResponse.setStatus(HttpServletResponse.SC_OK);
        httpServletResponse.getWriter().write(e.toString());
        httpServletResponse.getWriter().flush();
    }
  }

  public void doGet(HttpServletRequest httpServletRequest, HttpServletResponse httpServletResponse) throws ServletException, IOException {
      doPost(httpServletRequest, httpServletResponse);
  }
}
