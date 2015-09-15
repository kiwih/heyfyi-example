package heyfyi

import (
	"strings"

	"github.com/gocraft/web"
)

type URL string

const (
	HomeUrl                 URL = "/"
	ListFactUrl             URL = "/fact"
	CreateFactUrl           URL = "/fact/create"
	ViewFactUrl             URL = "/fact/view/:factId"
	DeleteFactUrl           URL = "/fact/delete/:factId"
	VoteOnFactUrl           URL = "/api/vote"
	ModerateFactUrl         URL = "/api/moderate"
	SignUpUrl               URL = "/signup"
	SignInUrl               URL = "/signin"
	SignOutUrl              URL = "/signout"
	VerificationUrl         URL = "/verify/:accountId/:verificationCode"
	RequestPasswordResetUrl URL = "/reset"
	ResetPasswordUrl        URL = "/reset/:accountId/:resetVerificationCode"
)

func (u URL) String() string {
	return string(u)
}

func (u URL) Make(param ...string) string {
	if len(param)%2 != 0 {
		panic("Make URL " + u.String() + " had non-even number of params")
	}

	retStr := u.String()

	for i := 0; i < len(param); i += 2 {
		retStr = strings.Replace(retStr, ":"+param[i], param[i+1], 1)
	}
	return retStr
}

func initRouter() *web.Router {

	rootRouter := web.New(Context{})
	rootRouter.Middleware(web.LoggerMiddleware)
	rootRouter.Middleware(web.ShowErrorsMiddleware)
	rootRouter.Middleware(web.StaticMiddleware("./files", web.StaticOption{Prefix: "/public"})) // "public" is a directory to serve files from.)
	rootRouter.Middleware((*Context).AssignStorageMiddleware)
	rootRouter.Middleware((*Context).AssignTemplatesAndSessionsMiddleware)
	rootRouter.Middleware((*Context).LoadUserMiddleware)
	rootRouter.Middleware((*Context).GetErrorMessagesMiddleware)
	rootRouter.Middleware((*Context).GetNotificationMessagesMiddleware)

	//rootRouter web paths
	rootRouter.Get(HomeUrl.String(), (*Context).HomeHandler)

	//sign up, sign in, etc
	rootRouter.Get(SignUpUrl.String(), (*Context).SignUpHandler)
	rootRouter.Post(SignUpUrl.String(), (*Context).DoSignUpHandler)
	rootRouter.Post(SignInUrl.String(), (*Context).DoSignInRequestHandler)
	rootRouter.Get(VerificationUrl.String(), (*Context).DoVerificationRequestHandler)

	//password reset handlers
	rootRouter.Get(RequestPasswordResetUrl.String(), (*Context).BeginPasswordResetRequestHandler)
	rootRouter.Post(RequestPasswordResetUrl.String(), (*Context).DoBeginPasswordResetRequestHandler)
	rootRouter.Get(ResetPasswordUrl.String(), (*Context).PasswordResetRequestHandler)
	rootRouter.Post(ResetPasswordUrl.String(), (*Context).DoPasswordResetRequestHandler)

	//viewing and listing facts handlers
	rootRouter.Get(ViewFactUrl.String(), (*Context).ViewFactHandler)
	rootRouter.Get(ListFactUrl.String(), (*Context).ListFactsHandler)

	//must be logged in for some handlers...
	loggedInRouter := rootRouter.Subrouter(LoggedInContext{}, "/")
	loggedInRouter.Middleware((*LoggedInContext).RequireAccountMiddleware)

	//sign out handler
	loggedInRouter.Post(SignOutUrl.String(), (*LoggedInContext).DoSignOutRequestHandler)

	//vote, moderate fact handlers
	loggedInRouter.Post(VoteOnFactUrl.String(), (*LoggedInContext).VoteOnFactHandler)
	loggedInRouter.Post(ModerateFactUrl.String(), (*LoggedInContext).ModerateFactHandler)

	//create, delete fact handlers
	loggedInRouter.Get(CreateFactUrl.String(), (*LoggedInContext).CreateFactHandler)
	loggedInRouter.Post(CreateFactUrl.String(), (*LoggedInContext).DoCreateFactHandler)
	loggedInRouter.Get(DeleteFactUrl.String(), (*LoggedInContext).DeleteFactHandler)
	loggedInRouter.Post(DeleteFactUrl.String(), (*LoggedInContext).DoDeleteFactHandler)

	return rootRouter
}
