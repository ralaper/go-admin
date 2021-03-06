package controller

import (
	"fmt"
	"github.com/GoAdminGroup/go-admin/context"
	"github.com/GoAdminGroup/go-admin/modules/auth"
	"github.com/GoAdminGroup/go-admin/modules/language"
	"github.com/GoAdminGroup/go-admin/modules/menu"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/constant"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/parameter"
	"github.com/GoAdminGroup/go-admin/plugins/admin/modules/table"
	"github.com/GoAdminGroup/go-admin/template"
	"github.com/GoAdminGroup/go-admin/template/types"
	"github.com/GoAdminGroup/go-admin/template/types/form"
	template2 "html/template"
	"net/http"
)

func ShowDetail(ctx *context.Context) {
	prefix := ctx.Query(constant.PrefixKey)
	id := ctx.Query(constant.DetailPKKey)
	panel := table.Get(prefix)
	user := auth.Auth(ctx)

	newPanel := panel.Copy()

	formModel := newPanel.GetForm()

	var fieldList types.FieldList

	if len(panel.GetDetail().FieldList) == 0 {
		fieldList = panel.GetInfo().FieldList
	} else {
		fieldList = panel.GetDetail().FieldList
	}

	formModel.FieldList = make([]types.FormField, len(fieldList))

	for i, field := range fieldList {
		formModel.FieldList[i] = types.FormField{
			Field:        field.Field,
			TypeName:     field.TypeName,
			Head:         field.Head,
			FormType:     form.Default,
			FieldDisplay: field.FieldDisplay,
		}
	}

	formData, _, _, _, _, err := newPanel.GetDataWithId(id)

	var alert template2.HTML

	if err != nil && alert == "" {
		alert = aAlert().SetTitle(template2.HTML(`<i class="icon fa fa-warning"></i> ` + language.Get("error") + `!`)).
			SetTheme("warning").
			SetContent(template2.HTML(err.Error())).
			GetContent()
	}

	paramStr := parameter.GetParam(ctx.Request.URL.Query(),
		panel.GetInfo().DefaultPageSize,
		panel.GetInfo().SortField,
		panel.GetInfo().GetSort()).GetRouteParamStr()

	editUrl := modules.AorEmpty(panel.GetEditable(), routePathWithPrefix("show_edit", prefix)+paramStr+
		"&"+constant.EditPKKey+"="+ctx.Query(constant.EditPKKey))
	deleteUrl := modules.AorEmpty(panel.GetDeletable(), routePathWithPrefix("delete", prefix)+paramStr)
	infoUrl := routePathWithPrefix("info", prefix) + paramStr

	editUrl = user.GetCheckPermissionByUrlMethod(editUrl, route("show_edit").Method())
	deleteUrl = user.GetCheckPermissionByUrlMethod(deleteUrl, route("delete").Method())

	deleteJs := ""

	if deleteUrl != "" {
		deleteJs = fmt.Sprintf(`<script>
function DeletePost(id) {
	swal({
			title: '%s',
			type: "warning",
			showCancelButton: true,
			confirmButtonColor: "#DD6B55",
			confirmButtonText: '%s',
			closeOnConfirm: false,
			cancelButtonText: '%s',
		},
		function () {
			$.ajax({
				method: 'post',
				url: '%s',
				data: {
					id: id
				},
				success: function (data) {
					if (typeof (data) === "string") {
						data = JSON.parse(data);
					}
					if (data.code === 200) {
						location.href = '%s'
					} else {
						swal(data.msg, '', 'error');
					}
				}
			});
		});
}

$('.delete-btn').on('click', function (event) {
	DeletePost(%s)
});

</script>`, language.Get("are you sure to delete"), language.Get("yes"), language.Get("cancel"), deleteUrl, infoUrl, id)
	}

	title := panel.GetDetail().Title

	if title == "" {
		title = panel.GetInfo().Title + language.Get("Detail")
	}

	desc := panel.GetDetail().Description

	if desc == "" {
		desc = panel.GetInfo().Description + language.Get("Detail")
	}

	tmpl, tmplName := aTemplate().GetTemplate(isPjax(ctx))
	buf := template.Execute(tmpl, tmplName, user, types.Panel{
		Content: alert + detailContent(aForm().
			SetTitle(template.HTML(title)).
			SetContent(formData).
			SetFooter(template.HTML(deleteJs)).
			SetInfoUrl(infoUrl).
			SetPrefix(config.PrefixFixSlash()), editUrl, deleteUrl),
		Description: desc,
		Title:       title,
	}, config, menu.GetGlobalMenu(user, conn).SetActiveClass(config.URLRemovePrefix(ctx.Path())))

	ctx.HTML(http.StatusOK, buf.String())
}
