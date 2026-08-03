package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"

	"ecommerce/internal/domain"
	"ecommerce/internal/middleware"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}

// newTestContext monta um *gin.Context isolado, sem depender de rotas nem de
// middlewares, para chamar o Handler diretamente — o nível mais barato de
// teste de Controller, no mesmo espírito do "Controller chamado diretamente"
// usado no projeto go-tests. body pode ser nil (GET/DELETE sem corpo).
func newTestContext(method, target string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)

	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	context.Request = request

	return context, recorder
}

// newJSONTestContext serializa payload como corpo JSON da requisição.
func newJSONTestContext(t *testing.T, method, target string, payload any) (*gin.Context, *httptest.ResponseRecorder) {
	t.Helper()

	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("serializar corpo da requisição: %v", err)
	}

	return newTestContext(method, target, body)
}

// setIDParam simula o parâmetro de rota "id" (:id) sem precisar de um
// gin.Engine com rotas registradas.
func setIDParam(context *gin.Context, id string) {
	context.Params = gin.Params{{Key: "id", Value: id}}
}

// setAuthenticatedUser simula o trabalho do middleware Authenticate,
// colocando o usuário autenticado no contexto exatamente como
// middleware.Authenticate faria após validar o cookie JWT.
func setAuthenticatedUser(context *gin.Context, user domain.AuthenticatedUser) {
	context.Set(middleware.AuthenticatedUserContextKey, &user)
}

// serve chama o Handler diretamente e força a gravação do header de
// resposta. Fora do gin.Engine (que é quem normalmente chama
// WriteHeaderNow após o handler retornar), c.Status(204) sozinho nunca
// chega a ser escrito no ResponseRecorder — só c.JSON/c.Data disparam a
// escrita implicitamente, por gravarem corpo.
func serve(context *gin.Context, handlerFunc gin.HandlerFunc) {
	handlerFunc(context)
	context.Writer.WriteHeaderNow()
}

// decodeJSONBody transforma o corpo gravado no ResponseRecorder no tipo
// esperado pelo cenário.
func decodeJSONBody[T any](t *testing.T, recorder *httptest.ResponseRecorder) T {
	t.Helper()

	var decoded T
	if err := json.Unmarshal(recorder.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decodificar corpo da resposta %q: %v", recorder.Body.String(), err)
	}

	return decoded
}
