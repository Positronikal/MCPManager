# Reporting Security Issues

1. On GitHub, navigate to the main page of the repository.

2. Under the repository name, click **Security**. If you cannot see the **Security** tab, select the  dropdown menu, and then click **Security**.

3. Click **Report a vulnerability** to open the advisory form.

4. Fill in the advisory details form.

  - In this form, only the title and description are mandatory. However, we recommend security researchers provide as much information as possible on the form so that the Positronikal team can make an informed decision about the submitted report.

  - For more information about the fields available and guidance on filling in the form, see [Best practices for writing repository security advisories](https://docs.github.com/en/code-security/security-advisories/guidance-on-reporting-and-writing-information-about-vulnerabilities/best-practices-for-writing-repository-security-advisories 'Best practices for writing repository security advisories').

5. At the bottom of the form, click **Submit report**. GitHub will display a message letting you know that maintainers have been notified and that you have a pending credit for this security advisory.

  - When the report is submitted, GitHub automatically adds the reporter of the vulnerability as a collaborator and as a credited user on the proposed advisory.

  - Optionally, click Start a temporary private fork if you want to start to fix the issue. Note that only the repository maintainer can merge changes from that private fork into the parent repository.

The next steps depend on the action taken by the repository maintainer.

## Sample Vulnerability Report
```
I identified potential security vulnerabilities in <product>.

I am committed to working with you to help resolve these issues. In
this report you will find everything you need to effectively coordinate
a resolution of these issues.

If at any point you have concerns or questions about this process,
please do not hesitate to reach out to me at <email>.

If you are NOT the correct point of contact for this report, please let
me know!

Summary
<Short summary of the problem. Make the impact and severity as clear as
possible. For example: An unsafe deserialization vulnerability allows
any unauthenticated user to execute arbitrary code on the server.>

Product
<product>

Tested Version
<version>

Details
<Give all details on the vulnerability. Pointing to the incriminated
source code is very helpful for the maintainer.>

PoC
<Complete instructions, including specific configuration details, to
reproduce the vulnerability.>

Impact
<impact>

Remediation
<Propose a remediation suggestion if you have one. Make it clear that
this is just a suggestion, as the maintainer might have a better idea
to fix the issue.>

GitHub Security Advisories
I understand that you may, at your discretion, create a private GitHub
Security Advisory and obtain CVE identification number from GitHub for
these findings. I also understand that you may also invite me to
collaborate and further discuss these findings in private before they
are published. If so, I agree to collaborate with you and review your
fix to make sure that all corner cases are covered. I understand that
GitHub usually reviews the request within 72 hours and the CVE details
will be published after you make your security advisory public. I
further understand that publishing a GitHub Security Advisory and a CVE
will help notify the downstream consumers of your project so they can
update to the fixed version, but isn't aleways necessary for every ptoject.

Credit
<List all researchers who contributed to this disclosure. If you found
the vulnerability with a specific tool, you can also credit this tool.>

Contact
<contact>
```

## Disclosure Policy
The Positronikal team is dedicated to correcting and disclosing vulnerabilities in any of our projects in order to protect users and ensure a coordinated disclosure with the vulnerability reporter. If the Positronikal team agrees that a reported vulnerability in our project poses a security risk, we will work with the reporting party to address the vulnerability in detail, and agree on the process for public disclosure.

The Positronikal team is also dedicated to working closely with the open source community and with projects that are upsteam of or depended on by our project that are affected by a vulnerability. When we identify a vulnerability in an upstream or dependency project, we will report it by contacting the publicly-listed security contact for the project if one exists; otherwise we will attempt to contact the project maintainers directly.

If the project team responds and agrees the issue poses a security risk, we will work with the project security team or maintainers to communicate the vulnerability in detail, and agree on the process for public disclosure. Responsibility for developing and releasing a patch for an upstream or dependency project lies firmly with that project's team, though we aim to facilitate this by providing detailed information about the vulnerability.

Our disclosure deadline for publicly disclosing a vulnerability is: 90 days after the first report to the project team.
