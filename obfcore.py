#!/usr/bin/env python3

import os
import sys
import json
import logging
from telegram import Update, InputFile
from telegram.ext import Application, CommandHandler, MessageHandler, filters, ContextTypes
from obfcore import BlankOBFv2  # Make sure obfcore.py exists

CONFIG_FILE = "config.json"
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

# Load Configuration
def load_config():
    if os.path.exists(CONFIG_FILE):
        with open(CONFIG_FILE, "r") as f:
            return json.load(f)
    return {}

# Save Configuration
def save_config(config):
    with open(CONFIG_FILE, "w") as f:
        json.dump(config, f, indent=4)

# /start Command
async def start(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    await update.message.reply_text(
        "👋 Welcome to Python Obfuscator Bot\n\n"
        "Send your Python code or .py file to get it obfuscated!\n"
        "Commands:\n"
        "/start - Show this message\n"
        "/recursive [N] - Set recursion level (default 1)\n"
        "/imports - Toggle including imports\n"
        "/obfuscate - Run obfuscation"
    )

# /recursive Command
async def set_recursive(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    user_id = update.message.from_user.id
    try:
        level = int(context.args[0]) if context.args else 1
        if level < 1:
            raise ValueError
        context.user_data[user_id] = context.user_data.get(user_id, {})
        context.user_data[user_id]["recursive"] = level
        await update.message.reply_text(f"🔁 Recursive obfuscation level set to {level}")
    except:
        await update.message.reply_text("⚠️ Usage: /recursive 2")

# /imports Command
async def toggle_imports(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    user_id = update.message.from_user.id
    context.user_data[user_id] = context.user_data.get(user_id, {})
    current = context.user_data[user_id].get("include_imports", False)
    context.user_data[user_id]["include_imports"] = not current
    status = "ON" if not current else "OFF"
    await update.message.reply_text(f"📦 Include imports is now {status}")

# Handle Python File
async def handle_file(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    user_id = update.message.from_user.id
    document = update.message.document
    if not document.file_name.endswith(".py"):
        await update.message.reply_text("⚠️ Only .py files are allowed.")
        return

    file = await document.get_file()
    path = f"/tmp/{user_id}.py"
    await file.download_to_drive(path)

    with open(path, "r") as f:
        context.user_data[user_id] = context.user_data.get(user_id, {})
        context.user_data[user_id]["code"] = f.read()

    os.remove(path)
    await update.message.reply_text("✅ Python file received. Use /obfuscate to process.")

# Handle Text
async def handle_text(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    user_id = update.message.from_user.id
    context.user_data[user_id] = context.user_data.get(user_id, {})
    context.user_data[user_id]["code"] = update.message.text
    await update.message.reply_text("📝 Code received. Use /obfuscate to process.")

# /obfuscate Command
async def obfuscate(update: Update, context: ContextTypes.DEFAULT_TYPE) -> None:
    user_id = update.message.from_user.id
    user_data = context.user_data.get(user_id, {})
    code = user_data.get("code")

    if not code:
        await update.message.reply_text("⚠️ No code to obfuscate.")
        return

    recursion = user_data.get("recursive", 1)
    include_imports = user_data.get("include_imports", False)

    try:
        obf = BlankOBFv2(code, include_imports, recursion)
        result = obf.obfuscate()

        if len(result) < 4000:
            await update.message.reply_text(f"```python\n{result}\n```", parse_mode="Markdown")
        else:
            file_path = f"obfuscated_{user_id}.py"
            with open(file_path, "w") as f:
                f.write(result)
            await update.message.reply_document(InputFile(file_path))
            os.remove(file_path)
    except Exception as e:
        logger.error("Obfuscation error: %s", e)
        await update.message.reply_text(f"❌ Obfuscation failed: {e}")

# Main Launcher
def main():
    config = load_config()

    if not config.get("token"):
        config["token"] = input("📲 Enter your Telegram Bot Token: ").strip()
        config["admin"] = input("👤 Enter your Admin Telegram ID: ").strip()
        save_config(config)
        print("✅ Configuration saved!")

    print("🚀 Bot is running...")
    app = Application.builder().token(config["token"]).build()

    app.add_handler(CommandHandler("start", start))
    app.add_handler(CommandHandler("recursive", set_recursive))
    app.add_handler(CommandHandler("imports", toggle_imports))
    app.add_handler(CommandHandler("obfuscate", obfuscate))
    app.add_handler(MessageHandler(filters.Document.FileExtension("py"), handle_file))
    app.add_handler(MessageHandler(filters.TEXT & ~filters.COMMAND, handle_text))

    app.run_polling()

if __name__ == "__main__":
    main()
