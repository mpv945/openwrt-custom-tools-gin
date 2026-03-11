$(function(){

    $("#loginForm").submit(function(e){

        e.preventDefault()

        let btn=$("#loginBtn")

        btn.prop("disabled",true)
        btn.text("Signing in...")

        $.ajax({

            url:"/api/login",

            type:"POST",

            data:$(this).serialize(),

            success:function(res){

                if(res.code===0){

                    window.location="/dashboard"

                }else{

                    showError(res.message)

                }

            },

            error:function(){

                showError("Server error")

            },

            complete:function(){

                btn.prop("disabled",false)

                btn.text("Sign In")

            }

        })

    })

})

function showError(msg){

    $("#alertBox")

        .removeClass("d-none")

        .text(msg)

}