package org.bsidessf.ctf.toolatte;

import java.io.IOException;
import jakarta.servlet.ServletException;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;

public class GenerateTokenServlet extends HttpServlet {
  public void doGet(HttpServletRequest httpServletRequest, HttpServletResponse httpServletResponse) throws ServletException, IOException {
    String response = null;
    try {
      response = TokenAPI.getGenerateResponse();

      httpServletResponse.setStatus(HttpServletResponse.SC_OK);
      httpServletResponse.getWriter().write(response);
      httpServletResponse.getWriter().flush();
    } catch(Exception e) {
      System.err.println("Error: " + e.toString());
      httpServletResponse.sendError(500);
    }
  }
}
